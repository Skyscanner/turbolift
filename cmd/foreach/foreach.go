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
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/skyscanner/turbolift/internal/logging"
)

var exec executor.Executor = executor.NewRealExecutor()

var (
	repoFile string = "repos.txt"
	helpFlag bool   = false
)

func parseForeachArgs(args []string) []string {
	strippedArgs := make([]string, 0)
MAIN:
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--repos":
			repoFile = args[i+1]
			i = i + 1
		case "--help":
			helpFlag = true
		default:
			// we've parsed everything that could be parsed; this is now the command
			strippedArgs = append(strippedArgs, args[i:]...)
			break MAIN
		}
	}

	return strippedArgs
}

func NewForeachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "foreach [flags] SHELL_COMMAND",
		Short:                 "Run a shell command against each working copy",
		Run:                   run,
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
	}

	// this flag will not be parsed (DisabledFlagParsing is on) but is here for the help context and auto complete
	cmd.Flags().StringVar(&repoFile, "repos", "repos.txt", "A file containing a list of repositories to clone.")

	return cmd
}

func run(c *cobra.Command, args []string) {
	logger := logging.NewLogger(c)

	/*
		Parsing is disabled for this command to make sure it doesn't capture flags from the subsequent command.
		E.g.: turbolift foreach ls -l   <- here, the -l would be captured by foreach, not by ls
		Because of this, we need a manual parsing of the arguments.
		Assumption is the foreach arguments will be parsed before the command and its arguments.
	*/
	args = parseForeachArgs(args)

	// check if the help flag was toggled
	if helpFlag {
		_ = c.Usage()
		return
	}

	readCampaignActivity := logger.StartActivity("Reading campaign data (%s)", repoFile)
	options := campaign.NewCampaignOptions()
	options.RepoFilename = repoFile
	dir, err := campaign.OpenCampaign(options)
	if err != nil {
		readCampaignActivity.EndWithFailure(err)
		return
	}
	readCampaignActivity.EndWithSuccess()

	var doneCount, skippedCount, errorCount int
	for _, repo := range dir.Repos {
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo
		command := strings.Join(args, " ")

		execActivity := logger.StartActivity("Executing %s in %s", command, repoDirPath)

		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			execActivity.EndWithWarningf("Directory %s does not exist - has it been cloned?", repoDirPath)
			skippedCount++
			continue
		}

		// Execute within a shell so that piping, redirection, etc are possible
		shellCommand := os.Getenv("SHELL")
		if shellCommand == "" {
			shellCommand = "sh"
		}
		shellArgs := []string{"-c", command}
		err := exec.Execute(execActivity.Writer(), repoDirPath, shellCommand, shellArgs...)

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
}
