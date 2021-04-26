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
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var gh github.GitHub = github.NewRealGitHub()
var g git.Git = git.NewRealGit()

func NewCreatePRsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-prs",
		Short: "Create pull requests for all repositories with changes",
		Run:   run,
	}

	return cmd
}

func run(c *cobra.Command, args []string) {
	dir, err := campaign.OpenCampaign()
	if err != nil {
		c.Printf(colors.Red("Error when reading campaign directory: %s\n"), err)
		return
	}

	doneCount := 0
	skippedCount := 0
	errorCount := 0
	for _, repo := range dir.Repos {
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo

		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			c.Printf(colors.Yellow("Not running against %s as the directory %s does not exist - has it been cloned?\n"), repo.FullRepoName, repoDirPath)
			skippedCount++
			continue
		}

		c.Println(repo.FullRepoName)

		c.Printf("Pushing changes in %s to origin\n", repoDirPath)
		err := g.Push(c.OutOrStdout(), repoDirPath, "origin", dir.Name)
		if err != nil {
			c.Printf(colors.Red("Error when pushing to upstream for %s: %s\n"), repo.FullRepoName, err)
			errorCount++
			continue
		}

		pullRequest := github.PullRequest{
			Title:        dir.PrTitle,
			Body:         dir.PrBody,
			UpstreamRepo: repo.FullRepoName,
		}
		c.Println("Creating PR")
		didCreate, err := gh.CreatePullRequest(c.OutOrStdout(), repoDirPath, pullRequest)

		if err != nil {
			c.Printf(colors.Red("Error when creating PR in %s: %s\n"), repo.FullRepoName, err)
			errorCount++
		} else if !didCreate {
			c.Printf(colors.Yellow("No PR created in %s\n"), repo.FullRepoName)
			skippedCount++
		} else {
			doneCount++
		}
	}

	if errorCount == 0 {
		c.Printf(colors.Green("✅ turbolift create-prs completed (%d OK, %d skipped)\n"), doneCount, skippedCount)
	} else {
		c.Printf(colors.Yellow("⚠️ turbolift create-prs completed with errors (%d OK, %d skipped, %d errored)\n"), doneCount, skippedCount, errorCount)
	}
}
