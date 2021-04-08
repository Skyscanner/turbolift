package clone

import (
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/skyscanner/turbolift/internal/simplelog"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var exec executor.Executor = executor.NewRealExecutor()

func CreateCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: " ", // TODO
		Run:   run,
	}

	return cmd
}

func run(c *cobra.Command, _ []string) {
	dir, err := campaign.OpenCampaignDirectory()
	if err != nil {
		c.Printf(simplelog.Red(c, "Error when reading campaign directory: %s\n", err))
		return
	}

	for _, repo := range dir.Repos {
		parentPath := path.Join("work", repo.OrgName)

		err := os.MkdirAll(parentPath, os.ModeDir|0755)
		if err != nil {
			c.Printf(simplelog.Red(c, "Error creating parent directory: %s: %s\n", parentPath, err))
		}

		// TODO: skip if the working copy is already cloned

		err = exec.Execute(parentPath, "gh", "repo", "fork", "--clone=true", repo.FullRepoName)
		if err != nil {
			c.Printf(simplelog.Red("Error when cloning %s: %s\n"), repo.FullRepoName, err)
		}

		// TODO: Implement branch creation
	}
}
