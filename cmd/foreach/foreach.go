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

package foreach

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/skyscanner/turbolift/internal/logging"

	"github.com/alessio/shellescape"
)

var exec executor.Executor = executor.NewRealExecutor()

var (
	repoFile string = "repos.txt"
)

func formatArguments(arguments []string) string {
	quotedArgs := make([]string, len(arguments))
	for i, arg := range arguments {
		quotedArgs[i] = shellescape.Quote(arg)
	}
	return strings.Join(quotedArgs, " ")
}

func NewForeachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "foreach [--repos REPOFILE] -- COMMAND [ARGUMENT...]",
		Short: "Run COMMAND against each working copy",
		Long:
`Run COMMAND against each working copy. Make sure to include a
double hyphen -- with space on both sides before COMMAND, as this
marks that no further options should be interpreted by turbolift.`,
		RunE: runE,
		Args: cobra.MinimumNArgs(1),
	}

	cmd.Flags().StringVar(&repoFile, "repos", "repos.txt", "A file containing a list of repositories to clone.")

	return cmd
}

func runE(c *cobra.Command, args []string) error {
	logger := logging.NewLogger(c)

	if c.ArgsLenAtDash() != 0 {
		return errors.New("Use -- to separate command")
	}

	readCampaignActivity := logger.StartActivity("Reading campaign data (%s)", repoFile)
	options := campaign.NewCampaignOptions()
	options.RepoFilename = repoFile
	dir, err := campaign.OpenCampaign(options)
	if err != nil {
		readCampaignActivity.EndWithFailure(err)
		return nil
	}
	readCampaignActivity.EndWithSuccess()

	// We shell escape these to avoid ambiguity in our logs, and give
	// the user something they could copy and paste.
	prettyArgs := formatArguments(args)

	var doneCount, skippedCount, errorCount int
	for _, repo := range dir.Repos {
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo

		execActivity := logger.StartActivity("Executing %s in %s", prettyArgs, repoDirPath)

		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			execActivity.EndWithWarningf("Directory %s does not exist - has it been cloned?", repoDirPath)
			skippedCount++
			continue
		}

		err := exec.Execute(execActivity.Writer(), repoDirPath, args[0], args[1:]...)

		if err != nil {
			execActivity.EndWithFailure(err)
			errorCount++
		} else {
			execActivity.EndWithSuccessAndEmitLogs()
			doneCount++
		}
	}

	if errorCount == 0 {
		logger.Successf("turbolift foreach completed %s(%s, %s)\n", colors.Normal(), colors.Green(doneCount, " OK"), colors.Yellow(skippedCount, " skipped"))
	} else {
		logger.Warnf("turbolift foreach completed with %s %s(%s, %s, %s)\n", colors.Red("errors"), colors.Normal(), colors.Green(doneCount, " OK"), colors.Yellow(skippedCount, " skipped"), colors.Red(errorCount, " errored"))
	}

	return nil
}
