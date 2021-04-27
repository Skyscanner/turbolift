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

package commit

import (
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var g git.Git = git.NewRealGit()

var message string

func NewCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Applies git commit -a -m '...' to all working copies, if they have changes",
		Run:   run,
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message to apply")
	err := cmd.MarkFlagRequired("message")
	if err != nil {
		panic(err)
	}

	return cmd
}

func run(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	readCampaignActivity := logger.StartActivity("Reading campaign data")
	dir, err := campaign.OpenCampaign()
	if err != nil {
		readCampaignActivity.EndWithFailure(err)
		return
	}
	readCampaignActivity.EndWithSuccess()

	doneCount := 0
	skippedCount := 0
	errorCount := 0
	for _, repo := range dir.Repos {
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo

		commitActivity := logger.StartActivity("Committing changes in %s", repo.FullRepoName)

		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			commitActivity.EndWithWarningf("Directory %s does not exist - has it been cloned?", repoDirPath)
			skippedCount++
			continue
		}

		isChanged, err := g.IsRepoChanged(commitActivity.Writer(), repoDirPath)
		if err != nil {
			commitActivity.EndWithFailure(err)
			errorCount++
			continue
		}

		if !isChanged {
			commitActivity.EndWithWarning("No changes - skipping commit")
			skippedCount++
			continue
		}

		err = g.Commit(commitActivity.Writer(), repoDirPath, message)
		if err != nil {
			commitActivity.EndWithFailure(err)
			errorCount++
		} else {
			commitActivity.EndWithSuccess()
			doneCount++
		}
	}

	if errorCount == 0 {
		logger.Successf("turbolift commit completed (%d OK, %d skipped)\n", doneCount, skippedCount)
	} else {
		logger.Warnf("turbolift commit completed with errors (%d OK, %d skipped, %d errored)\n", doneCount, skippedCount, errorCount)
	}
}
