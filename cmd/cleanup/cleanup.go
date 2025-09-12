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

package cleanup

import (
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var (
	gh          github.GitHub = github.NewRealGitHub()
	cleanupFile               = ".cleanup.txt"
	repoFile    string
)

func NewCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleans up forks used in this campaign",
		Run:   run,
	}

	cmd.Flags().StringVar(&repoFile, "repos", "repos.txt", "A file containing a list of repositories to cleanup.")

	return cmd
}

func run(c *cobra.Command, _ []string) {
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

	cleanupFileActivity := logger.StartActivity("Creating cleanup file (%s)", cleanupFile)
	deletableForks, err := os.Create(cleanupFile)
	if err != nil {
		cleanupFileActivity.EndWithFailure(err)
		return
	}
	cleanupFileActivity.EndWithSuccess()

	defer func(deletableForks *os.File) {
		err := deletableForks.Close()
		if err != nil {
			logger.Errorf("Error closing cleanup file: %s", err)
		}
	}(deletableForks)

	forksFound := false
	deletableForksFound := false
	var doneCount, errorCount, skippedCount int
	for _, repo := range dir.Repos {

		forkStatusActivity := logger.StartActivity("Checking whether %s is a fork", repo.FullRepoName)
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName)
		isFork, err := gh.IsFork(logger.Writer(), repoDirPath)
		if err != nil {
			errorCount++
			forkStatusActivity.EndWithFailure(err)
			continue
		}
		if !isFork {
			skippedCount++
			forkStatusActivity.EndWithSuccess()
			continue
		}
		forkStatusActivity.EndWithSuccess()

		forksFound = true

		prCheckActivity := logger.StartActivity("Checking for open PRs in %s", repo.FullRepoName)
		openUpstreamPR, err := gh.UserHasOpenUpstreamPRs(logger.Writer(), repo.FullRepoName)
		if err != nil {
			errorCount++
			prCheckActivity.EndWithFailure(err)
			continue
		}
		prCheckActivity.EndWithSuccess()
		if !openUpstreamPR {
			deletableForkActivity := logger.StartActivity("Adding fork of %s to cleanup file", repo.FullRepoName)
			originRepoName, err := gh.GetOriginRepoName(logger.Writer(), repoDirPath)
			if err != nil {
				errorCount++
				deletableForkActivity.EndWithFailure(err)
				continue
			}
			_, err = deletableForks.WriteString(originRepoName + "\n")
			if err != nil {
				errorCount++
				deletableForkActivity.EndWithFailure(err)
				continue
			}
			deletableForkActivity.EndWithSuccess()
			deletableForksFound = true
		}
		doneCount++
	}

	if errorCount == 0 {
		logger.Successf("turbolift cleanup completed %s(%s forks checked, %s non-forks skipped)\n", colors.Normal(), colors.Green(doneCount), colors.Yellow(skippedCount))
		if deletableForksFound {
			logger.Printf(" %s contains a list of forks used in this campaign that do not currently have an upstream PR open. Please check over these carefully. It is your responsibility to ensure that they are in fact to safe to delete.", cleanupFile)
			logger.Println("If you wish to delete these forks, run the following command:")
			logger.Printf("    for f in $(cat %s); do", cleanupFile)
			logger.Println("         gh repo delete --yes $f")
			logger.Println("         sleep 1")
			logger.Println("         done")
		} else {
			if forksFound {
				logger.Println("All forks used in this campaign appear to have an open upstream PR. No cleanup can be done at this time.")
			} else {
				logger.Println("No forks found in this campaign.")
			}
		}
	} else {
		logger.Errorf("turbolift cleanup completed with errors")
		logger.Warnf("turbolift cleanup completed with %s %s(%s forks checked, %s non-forks skipped, %s errored)\n", colors.Red("errors"), colors.Normal(), colors.Green(doneCount), colors.Yellow(skippedCount), colors.Red(errorCount))
		logger.Println("Please check errors above and fix if necessary")
	}
}
