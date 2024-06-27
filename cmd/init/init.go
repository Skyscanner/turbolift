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

package init

import (
	_ "embed"
	"github.com/skyscanner/turbolift/internal/campaign"
	"os"
	"path/filepath"

	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
)

var campaignName string

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Turbolift campaign directory",
		Run:   run,
	}

	cmd.Flags().StringVarP(&campaignName, "name", "n", "", "Campaign name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func run(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	createDirActivity := logger.StartActivity("Creating work directory")
	// Create a directory for both the campaign and its nested work directory
	workDirectory := filepath.Join(campaignName, "work")
	err := os.MkdirAll(workDirectory, os.ModeDir|0755)

	if err != nil {
		createDirActivity.EndWithFailure(err)
		return
	} else {
		createDirActivity.EndWithSuccess()
	}

	createFilesActivity := logger.StartActivity("Creating initial files")

	err = campaign.CreateInitialFiles(campaignName)
	if err != nil {
		createFilesActivity.EndWithFailure(err)
	}

	createFilesActivity.EndWithSuccess()

	logger.Successf("turbolift init is done - next:\n")
	logger.Println("\t1. Run", colors.Cyan("cd ", campaignName))
	logger.Println("\t2. Update", colors.Cyan("repos.txt"), "with the names of the repos that need changing (either manually or using a tool to generate a list of repos)")
	logger.Println("\t3. Run", colors.Cyan("turbolift clone"))
}
