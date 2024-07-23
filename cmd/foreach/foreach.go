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
	"fmt"
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
		Use:   "foreach [flags] -- COMMAND [ARGUMENT...]",
		Short: "Run COMMAND against each working copy",
		Long: `Run COMMAND against each working copy. Make sure to include a
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

	o := setupOutputFiles(dir.Name, prettyArgs)

	logger.Printf("Logs for all executions will be stored under %s", o.overallResultsDirectory)

	var doneCount, skippedCount, errorCount int
	for _, repo := range dir.Repos {
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo

		execActivity := logger.StartActivity("Executing { %s } in %s", prettyArgs, repoDirPath)

		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			execActivity.EndWithWarningf("Directory %s does not exist - has it been cloned?", repoDirPath)
			skippedCount++
			continue
		}

		err := exec.Execute(execActivity.Writer(), repoDirPath, args[0], args[1:]...)

		if err != nil {
			emitOutcomeToFiles(repo, o.failedReposFile, o.failedResultsDirectory, execActivity.Logs(), logger)
			execActivity.EndWithFailure(err)
			errorCount++
		} else {
			emitOutcomeToFiles(repo, o.successfulReposFile, o.successfulResultsDirectory, execActivity.Logs(), logger)
			execActivity.EndWithSuccessAndEmitLogs()
			doneCount++
		}
	}

	if errorCount == 0 {
		logger.Successf("turbolift foreach completed %s(%s, %s)\n", colors.Normal(), colors.Green(doneCount, " OK"), colors.Yellow(skippedCount, " skipped"))
	} else {
		logger.Warnf("turbolift foreach completed with %s %s(%s, %s, %s)\n", colors.Red("errors"), colors.Normal(), colors.Green(doneCount, " OK"), colors.Yellow(skippedCount, " skipped"), colors.Red(errorCount, " errored"))
	}

	logger.Printf("Logs for all executions have been stored under %s", o.overallResultsDirectory)
	logger.Printf("Names of successful repos have been written to %s", o.successfulReposFile.Name())
	logger.Printf("Names of failed repos have been written to %s", o.failedReposFile.Name())

	return nil
}

type outputLogFileDestinations struct {
	overallResultsDirectory string

	successfulResultsDirectory string
	successfulReposFile        *os.File

	failedResultsDirectory string
	failedReposFile        *os.File
}

// sets up a temporary directory to store success/failure logs etc
func setupOutputFiles(campaignName string, command string) outputLogFileDestinations {
	resultsDirectory, _ := os.MkdirTemp("", fmt.Sprintf("turbolift-foreach-%s-", campaignName))
	successfulResultsDirectory := path.Join(resultsDirectory, "successful")
	failedResultsDirectory := path.Join(resultsDirectory, "failed")
	_ = os.MkdirAll(successfulResultsDirectory, 0755)
	_ = os.MkdirAll(failedResultsDirectory, 0755)

	successfulReposTxt := path.Join(successfulResultsDirectory, "repos.txt")
	failedReposTxt := path.Join(failedResultsDirectory, "repos.txt")

	// create the files
	successfulReposFile, _ := os.Create(successfulReposTxt)
	failedReposFile, _ := os.Create(failedReposTxt)

	_, _ = successfulReposFile.WriteString(fmt.Sprintf("# This file contains the list of repositories that were successfully processed by turbolift foreach\n# for the command: %s\n", command))
	_, _ = failedReposFile.WriteString(fmt.Sprintf("# This file contains the list of repositories that failed to be processed by turbolift foreach\n# for the command: %s\n", command))

	return outputLogFileDestinations{
		overallResultsDirectory: resultsDirectory,

		successfulResultsDirectory: successfulResultsDirectory,
		successfulReposFile:        successfulReposFile,

		failedResultsDirectory: failedResultsDirectory,
		failedReposFile:        failedReposFile,
	}
}

func emitOutcomeToFiles(repo campaign.Repo, reposFile *os.File, logsDirectoryParent string, executionLogs string, logger *logging.Logger) {
	// write the repo name to the repos file
	_, err := reposFile.WriteString(repo.FullRepoName + "\n")
	if err != nil {
		logger.Errorf("Failed to write repo name to %s: %s", reposFile.Name(), err)
	}

	// write logs to a file under the logsParent directory, in a directory structure that mirrors that of the work directory
	logsDir := path.Join(logsDirectoryParent, repo.FullRepoName)
	logsFile := path.Join(logsDir, "logs.txt")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		logger.Errorf("Failed to create directory %s: %s", logsDir, err)
	}

	logs, _ := os.Create(logsFile)
	_, err = logs.WriteString(executionLogs)
	if err != nil {
		logger.Errorf("Failed to write logs to %s: %s", logsFile, err)
	}
}
