/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
)

var (
	gh github.GitHub = github.NewRealGitHub()
	g  git.Git       = git.NewRealGit()
)

var (
	nofork   bool
	repoFile string
)

func NewCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone all repositories",
		Run:   run,
	}

	cmd.Flags().BoolVar(&nofork, "no-fork", false, "Will not fork, just clone and create a branch.")
	cmd.Flags().StringVar(&repoFile, "repos", "repos.txt", "A file containing a list of repositories to clone.")

	return cmd
}

func run(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	readCampaignActivity := logger.StartActivity("Reading campaign data")
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
		orgDirPath := path.Join("work", repo.OrgName) // i.e. work/org

		var cloneActivity *logging.Activity
		if nofork {
			cloneActivity = logger.StartActivity("Cloning %s into %s/%s", repo.FullRepoName, orgDirPath, repo.RepoName)
		} else {
			cloneActivity = logger.StartActivity("Forking and cloning %s into %s/%s", repo.FullRepoName, orgDirPath, repo.RepoName)
		}

		err := os.MkdirAll(orgDirPath, os.ModeDir|0755)
		if err != nil {
			cloneActivity.EndWithFailuref("Unable to create org directory: %s", err)
			errorCount++
			break
		}

		repoDirPath := path.Join(orgDirPath, repo.RepoName) // i.e. work/org/repo
		// skip if the working copy is already cloned
		if _, err = os.Stat(repoDirPath); !os.IsNotExist(err) {
			cloneActivity.EndWithWarningf("Directory already exists")
			skippedCount++
			continue
		}

		if nofork {
			err = gh.Clone(cloneActivity.Writer(), orgDirPath, repo.FullRepoName)
		} else {
			err = gh.ForkAndClone(cloneActivity.Writer(), orgDirPath, repo.FullRepoName)
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
