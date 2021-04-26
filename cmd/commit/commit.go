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
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/git"
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

		isChanged, err := g.IsRepoChanged(c.OutOrStdout(), repoDirPath)
		if err != nil {
			c.Printf(colors.Red("Error when checking for changes in %s: %s\n"), repo.FullRepoName, err)
			errorCount++
			continue
		}

		if !isChanged {
			c.Printf(colors.Yellow("No changes in %s - skipping commit\n"), repo.FullRepoName)
			skippedCount++
			continue
		}

		c.Printf("Committing changes in %s\n", repo.FullRepoName)

		err = g.Commit(c.OutOrStdout(), repoDirPath, message)
		if err != nil {
			c.Printf(colors.Red("Error when committing changes in %s: %s\n"), repo.FullRepoName, err)
			errorCount++
		} else {
			doneCount++
		}
	}

	if errorCount == 0 {
		c.Printf(colors.Green("✅ turbolift commit completed (%d OK, %d skipped)\n"), doneCount, skippedCount)
	} else {
		c.Printf(colors.Yellow("⚠️ turbolift commit completed with errors (%d OK, %d skipped, %d errored)\n"), doneCount, skippedCount, errorCount)
	}
}
