package foreach

import (
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
)

var exec executor.Executor = executor.NewRealExecutor()

func NewForeachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "foreach -- SHELL_COMMAND",
		Short: "Run a shell command against each working copy",
		Run:   run,
		Args:  cobra.MinimumNArgs(1),
	}

	return cmd
}

func run(c *cobra.Command, args []string) {
	dir, err := campaign.OpenCampaign()
	if err != nil {
		c.Printf(colors.Red("Error when reading campaign directory: %s\n"), err)
		return
	}

	var doneCount, skippedCount, errorCount int
	for _, repo := range dir.Repos {
		repoDirPath := path.Join("work", repo.OrgName, repo.RepoName) // i.e. work/org/repo

		// skip if the working copy does not exist
		if _, err = os.Stat(repoDirPath); os.IsNotExist(err) {
			c.Printf(colors.Yellow("Not running against %s as the directory %s does not exist - has it been cloned?\n"), repo.FullRepoName, repoDirPath)
			skippedCount++
			continue
		}

		c.Println(repo.FullRepoName)

		// Execute within a shell so that piping, redirection, etc are possible
		shellCommand := os.Getenv("SHELL")
		if shellCommand == "" {
			shellCommand = "sh"
		}
		shellArgs := []string{"-c", strings.Join(args, " ")}
		err := exec.Execute(c.OutOrStdout(), repoDirPath, shellCommand, shellArgs...)

		if err != nil {
			c.Printf(colors.Red("Error when executing command in %s: %s\n"), repo.FullRepoName, err)
			errorCount++
		} else {
			doneCount++
		}
	}

	if errorCount == 0 {
		c.Printf(colors.Green("✅ turbolift foreach completed (%d OK, %d skipped)\n"), doneCount, skippedCount)
	} else {
		c.Printf(colors.Yellow("⚠️ turbolift foreach completed with errors (%d OK, %d skipped, %d errored)\n"), doneCount, skippedCount, errorCount)
	}
}
