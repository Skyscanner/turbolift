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

package create_prs

import (
	"os"
	"path"
	"time"

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
	isDraft  bool
	repoFile string
	sleep    time.Duration
)

func NewCreatePRsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-prs",
		Short: "Create pull requests for all repositories with changes",
		Run:   run,
	}

	cmd.Flags().DurationVar(&sleep, "sleep", 0, "Fixed sleep in between PR creations (to spread load on CI infrastructure)")
	cmd.Flags().BoolVar(&isDraft, "draft", false, "Creates the Pull Request as Draft PR")
	cmd.Flags().StringVar(&repoFile, "repos", "repos.txt", "A file containing a list of repositories to clone.")

	return cmd
}

func run(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	readCampaignActivity := logger.StartActivity("Reading campaign data")
	defaultOptions := campaign.NewCampaignOptions()
	defaultOptions.RepoFilename = repoFile
	dir, err := campaign.OpenCampaign(defaultOptions)
	if err != nil {
		readCampaignActivity.EndWithFailure(err)
		return
	}
	readCampaignActivity.EndWithSuccess()

	doneCount := 0
	skippedCount := 0
	errorCount := 0
	for _, repo := range dir.Repos {
		if sleep > 0 {
			logger.Successf("Sleeping for %s", sleep)
			time.Sleep(sleep)
		}

		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo

		pushActivity := logger.StartActivity("Pushing changes in %s to origin", repo.FullRepoName)
		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			pushActivity.EndWithWarningf("Directory %s does not exist - has it been cloned?", repoDirPath)
			skippedCount++
			continue
		}

		err := g.Push(pushActivity.Writer(), repoDirPath, "origin", dir.Name)
		if err != nil {
			pushActivity.EndWithFailure(err)
			errorCount++
			continue
		}
		pushActivity.EndWithSuccess()

		var createPrActivity *logging.Activity
		if isDraft {
			createPrActivity = logger.StartActivity("Creating Draft PR in %s", repo.FullRepoName)
		} else {
			createPrActivity = logger.StartActivity("Creating PR in %s", repo.FullRepoName)
		}

		pullRequest := github.PullRequest{
			Title:        dir.PrTitle,
			Body:         dir.PrBody,
			UpstreamRepo: repo.FullRepoName,
			IsDraft:      isDraft,
		}

		didCreate, err := gh.CreatePullRequest(createPrActivity.Writer(), repoDirPath, pullRequest)

		if err != nil {
			createPrActivity.EndWithFailure(err)
			errorCount++
		} else if !didCreate {
			createPrActivity.EndWithWarningf("No PR created in %s", repo.FullRepoName)
			skippedCount++
		} else {
			createPrActivity.EndWithSuccess()
			doneCount++
		}
	}

	if errorCount == 0 {
		logger.Successf("turbolift create-prs completed %s(%s, %s)\n", colors.Normal(), colors.Green(doneCount, " OK"), colors.Yellow(skippedCount, " skipped"))
	} else {
		logger.Warnf("turbolift create-prs completed with %s %s(%s, %s, %s)\n", colors.Red("errors"), colors.Normal(), colors.Green(doneCount, " OK"), colors.Yellow(skippedCount, " skipped"), colors.Red(errorCount, " errored"))
	}
}
