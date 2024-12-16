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

package create_prs

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/skyscanner/turbolift/internal/prompt"

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
	p  prompt.Prompt = prompt.NewRealPrompt()
)

var (
	isDraft           bool
	repoFile          string
	prDescriptionFile string
	sleep             time.Duration
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
	cmd.Flags().StringVar(&prDescriptionFile, "description", "README.md", "A file containing the title and description for the PRs.")

	return cmd
}

func run(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	readCampaignActivity := logger.StartActivity("Reading campaign data (%s, %s)", repoFile, prDescriptionFile)
	options := campaign.NewCampaignOptions()
	options.RepoFilename = repoFile
	options.PrDescriptionFilename = prDescriptionFile
	dir, err := campaign.OpenCampaign(options)
	if err != nil {
		readCampaignActivity.EndWithFailure(err)
		return
	}
	readCampaignActivity.EndWithSuccess()

	// checking whether the description has changed
	if prDescriptionUnchanged(dir) {
		if !p.AskConfirm(fmt.Sprintf("It looks like the PR title and/or description may not have been updated in %s. Are you sure you want to proceed?", prDescriptionFile)) {
			return
		}
	}

	doneCount := 0
	skippedCount := 0
	errorCount := 0
	for i, repo := range dir.Repos {
		if i > 0 && sleep > 0 {
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

func prDescriptionUnchanged(dir *campaign.Campaign) bool {
	originalPrTitleTodo := "TODO: Title of Pull Request"
	originalPrBodyTodo := "TODO: This file will serve as both a README and the description of the PR."
	return strings.Contains(dir.PrTitle, originalPrTitleTodo) || strings.Contains(dir.PrBody, originalPrBodyTodo) || dir.PrTitle == ""
}
