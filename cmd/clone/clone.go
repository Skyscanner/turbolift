/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * https://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package clone

import (
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/logging"
)

var (
	gh github.GitHub = github.NewRealGitHub()
	g  git.Git       = git.NewRealGit()
)

var (
	forceFork bool
	repoFile  string
	prsFile   string
)

func NewCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone all repositories",
		Run:   run,
	}

	cmd.Flags().BoolVar(&forceFork, "fork", false, "Force forking, instead of turbolift choosing whether to fork/branch based on permissions")
	cmd.Flags().StringVar(&repoFile, "repos", "repos.txt", "A file containing a list of repositories to clone.")
	cmd.Flags().StringVar(&prsFile, "from-prs", "",
		"A file containing PR URLs or org/repo#N shorthand to assimilate. "+
			"Each PR's head branch is checked out locally and recorded as a "+
			"branch annotation in repos.txt. Mutually exclusive with --repos.")

	return cmd
}

func run(c *cobra.Command, args []string) {
	logger := logging.NewLogger(c)

	if prsFile != "" {
		// Mutual exclusion: honouring both --from-prs and a user-specified
		// --repos file would have ambiguous semantics (which list wins?), so
		// reject early. --repos stays on its default value for the PR flow
		// because we write into the default repos.txt.
		if c.Flags().Changed("repos") {
			logger.Errorf("--from-prs and --repos are mutually exclusive")
			return
		}
		runFromPRs(c, args)
		return
	}
	runNormal(c, args)
}

func runNormal(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	readCampaignActivity := logger.StartActivity("Reading campaign data (%s)", repoFile)
	options := campaign.NewCampaignOptions()
	options.RepoFilename = repoFile
	dir, err := campaign.OpenCampaign(options)
	if err != nil {
		readCampaignActivity.EndWithFailure(err)
		return
	}
	readCampaignActivity.EndWithSuccess()

	var doneCount, skippedCount, errorCount int
	for _, repo := range dir.Repos {
		orgDirPath := path.Join("work", repo.OrgName)       // i.e. work/org
		repoDirPath := path.Join(orgDirPath, repo.RepoName) // i.e. work/org/repo

		var cloneActivity *logging.Activity

		// Determine whether we need to fork or clone
		var fork bool

		if forceFork {
			fork = true
		} else {
			res, err := gh.IsPushable(logger.Writer(), repo.FullRepoName)
			if err != nil {
				logger.Warnf("Unable to determine if we can push to %s: %s", repo.FullRepoName, err)
				fork = true
			} else {
				fork = !res
			}
		}

		if fork {
			cloneActivity = logger.StartActivity("Forking and cloning %s into %s/%s", repo.FullRepoName, orgDirPath, repo.RepoName)
		} else {
			cloneActivity = logger.StartActivity("Cloning %s into %s/%s", repo.FullRepoName, orgDirPath, repo.RepoName)
		}

		err := os.MkdirAll(orgDirPath, os.ModeDir|0o755)
		if err != nil {
			cloneActivity.EndWithFailuref("Unable to create org directory: %s", err)
			errorCount++
			break
		}

		// skip if the working copy is already cloned
		if _, err = os.Stat(repoDirPath); !os.IsNotExist(err) {
			cloneActivity.EndWithWarningf("Directory already exists")
			skippedCount++
			continue
		}

		if fork {
			err = gh.ForkAndClone(cloneActivity.Writer(), orgDirPath, repo.FullRepoName)
		} else {
			err = gh.Clone(cloneActivity.Writer(), orgDirPath, repo.FullRepoName)
		}

		if err != nil {
			cloneActivity.EndWithFailure(err)
			errorCount++
			continue
		}

		cloneActivity.EndWithSuccess()

		createBranchActivity := logger.StartActivity("Creating branch %s in %s", dir.Name, repo.FullRepoName)

		err = g.Checkout(createBranchActivity.Writer(), repoDirPath, dir.Name)
		if err != nil {
			createBranchActivity.EndWithFailure(err)
			errorCount++
			continue
		}
		createBranchActivity.EndWithSuccess()

		if fork {
			pullFromUpstreamActivity := logger.StartActivity("Pulling latest changes from %s", repo.FullRepoName)
			var defaultBranch string
			defaultBranch, err = gh.GetDefaultBranchName(pullFromUpstreamActivity.Writer(), repoDirPath, repo.FullRepoName)
			if err != nil {
				pullFromUpstreamActivity.EndWithFailure(err)
				errorCount++
				continue
			}
			err = g.Pull(pullFromUpstreamActivity.Writer(), repoDirPath, "upstream", defaultBranch)
			if err != nil {
				pullFromUpstreamActivity.EndWithFailure(err)
				logger.Printf("\nWe weren't able to pull the latest upstream changes into your fork of %s. This is probably because you have a pre-existing fork with commits ahead of upstream. Please change this or delete your fork, and try again.\n", repo.FullRepoName)
				errorCount++
				continue
			}
			pullFromUpstreamActivity.EndWithSuccess()
		}

		doneCount++
	}

	if errorCount == 0 {
		logger.Successf("turbolift clone completed %s(%s repos cloned, %s repos skipped)\n", colors.Normal(), colors.Green(doneCount), colors.Yellow(skippedCount))
	} else {
		logger.Warnf("turbolift clone completed with %s %s(%s repos cloned, %s repos skipped, %s repos errored)\n", colors.Red("errors"), colors.Normal(), colors.Green(doneCount), colors.Yellow(skippedCount), colors.Red(errorCount))
		logger.Println("Please check errors above and fix if necessary")
	}
	logger.Println("To continue:")
	logger.Println("\t1. Make your changes in the cloned repositories within the", colors.Cyan("work"), "directory")
	logger.Println("\t2. Add new files across all repos using", colors.Cyan(`turbolift foreach -- git add -A`))
	logger.Println("\t3. Commit changes across all repos using", colors.Cyan(`turbolift commit --message "Your commit message"`))
	logger.Println("\t4. Change the PR title and description in the", colors.Cyan(`README.md`), "of a campaign")
}

// runFromPRs implements `clone --from-prs`. It reads PR refs from prsFile,
// clones each repo, runs `gh pr checkout` to land the PR's head branch, and
// then records those branches as annotations in repos.txt. We do all clones
// first and the repos.txt write last — this means a conflict detected by
// UpsertBranchAnnotations leaves repos.txt untouched rather than half-written.
func runFromPRs(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	readActivity := logger.StartActivity("Reading PRs file (%s)", prsFile)
	prs, err := campaign.ReadPRsFile(prsFile)
	if err != nil {
		readActivity.EndWithFailure(err)
		return
	}
	readActivity.EndWithSuccess()

	var doneCount, skippedCount, errorCount int
	collectedBranches := map[string]string{}

	for _, pr := range prs {
		orgDirPath := path.Join("work", pr.OrgName)
		repoDirPath := path.Join(orgDirPath, pr.RepoName)

		// The identifier we pass to `gh` for clone/fork/push permission checks.
		// For GHE PRs we include the host so `gh` hits the right instance.
		cloneTarget := pr.OrgName + "/" + pr.RepoName
		if pr.Host != "" {
			cloneTarget = pr.Host + "/" + cloneTarget
		}

		activity := logger.StartActivity("Assimilating PR #%d for %s", pr.Number, cloneTarget)

		if err := os.MkdirAll(orgDirPath, os.ModeDir|0o755); err != nil {
			activity.EndWithFailuref("Unable to create org directory: %s", err)
			errorCount++
			continue
		}

		// Skip-if-present mirrors the normal clone flow. This makes retries
		// safe: a user who hits a conflict, fixes repos.txt, and re-runs
		// should not see clones redone or errors for already-done work.
		if _, err := os.Stat(repoDirPath); !os.IsNotExist(err) {
			activity.EndWithWarningf("Directory already exists")
			skippedCount++
			// Still capture the current branch so UpsertBranchAnnotations
			// can reconcile. Key by cloneTarget (not FullRepoName) so GHE
			// repos.txt entries like `host/org/repo` match — FullRepoName
			// deliberately strips the host.
			if b, bErr := g.GetCurrentBranchName(activity.Writer(), repoDirPath); bErr == nil {
				collectedBranches[cloneTarget] = b
			}
			continue
		}

		// Decide fork vs direct clone using the same permission check as
		// the normal flow. In practice assimilation usually implies direct
		// clone (the PR author has push access) but honouring --fork keeps
		// behaviour consistent.
		var fork bool
		if forceFork {
			fork = true
		} else {
			res, permErr := gh.IsPushable(logger.Writer(), cloneTarget)
			if permErr != nil {
				logger.Warnf("Unable to determine if we can push to %s: %s", cloneTarget, permErr)
				fork = true
			} else {
				fork = !res
			}
		}

		if fork {
			err = gh.ForkAndClone(activity.Writer(), orgDirPath, cloneTarget)
		} else {
			err = gh.Clone(activity.Writer(), orgDirPath, cloneTarget)
		}
		if err != nil {
			activity.EndWithFailure(err)
			errorCount++
			continue
		}

		// `gh pr checkout` fetches the PR head and checks it out. For PRs
		// from forks it also configures a remote and sets upstream tracking
		// — that's why we use it instead of a hand-rolled git fetch+checkout.
		if err := gh.CheckoutPR(activity.Writer(), repoDirPath, pr.Number); err != nil {
			activity.EndWithFailure(err)
			errorCount++
			continue
		}

		// Capture the checked-out branch so we can record it in repos.txt.
		// Key by cloneTarget (not FullRepoName) so GHE repos.txt entries like
		// `host/org/repo` match the right line when UpsertBranchAnnotations
		// looks them up.
		branch, bErr := g.GetCurrentBranchName(activity.Writer(), repoDirPath)
		if bErr != nil {
			activity.EndWithFailure(bErr)
			errorCount++
			continue
		}
		collectedBranches[cloneTarget] = branch

		activity.EndWithSuccess()
		doneCount++
	}

	// Single atomic write. If any repo has a conflicting annotation already,
	// this errors without touching the file and we surface the failure so the
	// user can resolve manually before re-running.
	upsertActivity := logger.StartActivity("Updating repos.txt with PR branch annotations")
	if err := campaign.UpsertBranchAnnotations("repos.txt", collectedBranches); err != nil {
		upsertActivity.EndWithFailure(err)
		logger.Warnf("repos.txt was not modified. Resolve the conflict above and re-run.")
	} else {
		upsertActivity.EndWithSuccess()
	}

	if errorCount == 0 {
		logger.Successf("turbolift clone completed %s(%s repos cloned, %s repos skipped)\n",
			colors.Normal(), colors.Green(doneCount), colors.Yellow(skippedCount))
	} else {
		logger.Warnf("turbolift clone completed with %s %s(%s repos cloned, %s repos skipped, %s repos errored)\n",
			colors.Red("errors"), colors.Normal(),
			colors.Green(doneCount), colors.Yellow(skippedCount), colors.Red(errorCount))
	}
	logger.Println("To continue:")
	logger.Println("\t1. Make your changes in the cloned repositories within the", colors.Cyan("work"), "directory")
	logger.Println("\t2. Commit changes across all repos using", colors.Cyan(`turbolift commit --message "Your commit message"`))
	logger.Println("\t3. Push the updated PR branches using", colors.Cyan(`turbolift update-prs --push`))
}
