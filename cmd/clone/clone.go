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
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var gh github.GitHub = github.NewRealGitHub()
var g git.Git = git.NewRealGit()

func NewCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone all repositories",
		Run:   run,
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

	var doneCount, skippedCount, errorCount int
	for _, repo := range dir.Repos {
		orgDirPath := path.Join("work", repo.OrgName) // i.e. work/org

		forkCloneActivity := logger.StartActivity("Forking and cloning %s into %s/%s", repo.FullRepoName, orgDirPath, repo.RepoName)

		err := os.MkdirAll(orgDirPath, os.ModeDir|0755)
		if err != nil {
			forkCloneActivity.EndWithFailuref("Unable to create org directory: %s", err)
			errorCount++
			break
		}

		repoDirPath := path.Join(orgDirPath, repo.RepoName) // i.e. work/org/repo
		// skip if the working copy is already cloned
		if _, err = os.Stat(repoDirPath); !os.IsNotExist(err) {
			forkCloneActivity.EndWithWarningf("Directory already exists")
			skippedCount++
			continue
		}

		err = gh.ForkAndClone(forkCloneActivity.Writer(), orgDirPath, repo.FullRepoName)
		if err != nil {
			forkCloneActivity.EndWithFailure(err)
			errorCount++
			continue
		}

		forkCloneActivity.EndWithSuccess()

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
		logger.Successf("turbolift clone completed (%d repos cloned, %d repos skipped)\n", doneCount, skippedCount)
	} else {
		logger.Warnf("turbolift clone completed with errors (%d repos cloned, %d repos skipped, %d repos errored)\n", doneCount, skippedCount, errorCount)
		logger.Println("Please check errors above and fix if necessary")
	}
	logger.Println("To continue:")
	logger.Println("1. Make your changes in the cloned repositories within the", colors.Cyan("work"), "directory")
	logger.Println("2. Commit changes across all repos using", colors.Cyan("turbolift commit --message \"Your commit message\""))
}
