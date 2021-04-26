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
	dir, err := campaign.OpenCampaign()
	if err != nil {
		c.Printf(colors.Red("Error when reading campaign directory: %s\n"), err)
		return
	}

	var doneCount, skippedCount, errorCount int
	for _, repo := range dir.Repos {
		orgDirPath := path.Join("work", repo.OrgName) // i.e. work/org

		err := os.MkdirAll(orgDirPath, os.ModeDir|0755)
		if err != nil {
			c.Printf(colors.Red("Error creating parent directory: %s: %s\n"), orgDirPath, err)
			errorCount++
			break
		}

		c.Printf(colors.Cyan("== %s =>\n"), repo.FullRepoName)

		repoDirPath := path.Join(orgDirPath, repo.RepoName) // i.e. work/org/repo
		// skip if the working copy is already cloned
		if _, err = os.Stat(repoDirPath); !os.IsNotExist(err) {
			c.Printf(colors.Yellow("Not cloning %s as a directory already exists at %s\n"), repo.FullRepoName, repoDirPath)
			skippedCount++
			continue
		}

		c.Printf("Forking and cloning %s into %s/%s\n", repo.FullRepoName, orgDirPath, repo.RepoName)
		err = gh.ForkAndClone(c.OutOrStdout(), orgDirPath, repo.FullRepoName)
		if err != nil {
			c.Printf(colors.Red("Error when cloning %s: %s\n"), repo.FullRepoName, err)
			errorCount++
			continue
		}

		c.Printf("Creating branch %s in %s/%s\n", dir.Name, orgDirPath, repo.RepoName)
		err = g.Checkout(c.OutOrStdout(), repoDirPath, dir.Name)
		if err != nil {
			c.Printf(colors.Red("Error when creating branch: %s\n"), err)
			errorCount++
			continue
		}
		doneCount++
	}

	if errorCount == 0 {
		c.Printf(colors.Green("✅ turbolift clone completed (%d repos cloned, %d repos skipped)\n"), doneCount, skippedCount)
	} else {
		c.Printf(colors.Yellow("⚠️ turbolift clone completed with errors (%d repos cloned, %d repos skipped, %d repos errored)\n"), doneCount, skippedCount, errorCount)
		c.Println("Please check errors above and fix if necessary")
	}
	c.Println("To continue:")
	c.Println("1. Make your changes in the cloned repositories within the", colors.Cyan("work"), "directory")
	c.Println("2. Commit changes across all repos using", colors.Cyan("turbolift commit --message \"Your commit message\""))
}
