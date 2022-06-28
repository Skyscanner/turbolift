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

package pr_status

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
)

var reactionsOrder = []string{
	"THUMBS_UP",
	"THUMBS_DOWN",
	"LAUGH",
	"HOORAY",
	"CONFUSED",
	"HEART",
	"ROCKET",
	"EYES",
}

var reactionsMapping = map[string]string{
	"THUMBS_UP":   "👍",
	"THUMBS_DOWN": "👎",
	"LAUGH":       "😆",
	"HOORAY":      "🎉",
	"CONFUSED":    "😕",
	"HEART":       "❤️",
	"ROCKET":      "🚀",
	"EYES":        "👀",
}

var gh github.GitHub = github.NewRealGitHub()

var list bool

func NewPrStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr-status",
		Short: "Displays the status of PRs",
		Run:   run,
	}
	cmd.Flags().BoolVar(&list, "list", false, "Displays a listing by PR")

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

	statuses := make(map[string]int)
	reactions := make(map[string]int)

	detailsTable := table.New("Repository", "State", "Reviews", "URL")
	detailsTable.WithHeaderFormatter(color.New(color.Underline).SprintfFunc())
	detailsTable.WithFirstColumnFormatter(color.New(color.FgCyan).SprintfFunc())
	detailsTable.WithWriter(logger.Writer())

	for _, repo := range dir.Repos {
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo

		checkStatusActivity := logger.StartActivity("Checking PR status for %s", repo.FullRepoName)

		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			checkStatusActivity.EndWithWarningf("Directory %s does not exist - has it been cloned?", repoDirPath)
			statuses["SKIPPED"]++
			continue
		}

		prStatus, err := gh.GetPR(checkStatusActivity.Writer(), repoDirPath, dir.Name)
		if err != nil {
			checkStatusActivity.EndWithFailuref("No PR found: %v", err)
			statuses["NO_PR"]++
			continue
		}

		statuses[prStatus.State]++

		for _, reaction := range prStatus.ReactionGroups {
			reactions[reaction.Content] += reaction.Users.TotalCount
		}

		detailsTable.AddRow(repo.FullRepoName, prStatus.State, prStatus.ReviewDecision, prStatus.Url)

		checkStatusActivity.EndWithSuccess()
	}

	logger.Successf("turbolift pr-status completed\n")

	logger.Println()

	if list {
		detailsTable.Print()
		logger.Println()
	}

	summaryTable := table.New("State", "Count")
	summaryTable.WithHeaderFormatter(color.New(color.Underline).SprintfFunc())
	summaryTable.WithFirstColumnFormatter(color.New(color.FgCyan).SprintfFunc())
	summaryTable.WithWriter(logger.Writer())

	summaryTable.AddRow("Merged", statuses["MERGED"])
	summaryTable.AddRow("Open", statuses["OPEN"])
	summaryTable.AddRow("Closed", statuses["CLOSED"])
	summaryTable.AddRow("Skipped", statuses["SKIPPED"])
	summaryTable.AddRow("No PR Found", statuses["NO_PR"])

	summaryTable.Print()

	logger.Println()

	anyReactionsToShow := false
	var reactionsOutput []string
	for _, key := range reactionsOrder {
		if reactions[key] > 0 {
			reactionsOutput = append(reactionsOutput, fmt.Sprintf("%s %d", reactionsMapping[key], reactions[key]))
			anyReactionsToShow = true
		}
	}
	if anyReactionsToShow {
		logger.Println("Reactions:", strings.Join(reactionsOutput, "   "))
	}
}
