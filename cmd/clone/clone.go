package clone

import (
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/executor"
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
		c.Printf(colors.Red("Error when reading campaign directory: %s\n"), err)
		return
	}

	for _, repo := range dir.Repos {
		parentPath := path.Join("work", repo.OrgName)

		err := os.MkdirAll(parentPath, os.ModeDir|0755)
		if err != nil {
			c.Printf(colors.Red("Error creating parent directory: %s: %s\n"), parentPath, err)
			break
		}

		workingCopyPath := path.Join(parentPath, repo.RepoName)
		// skip if the working copy is already cloned
		if _, err = os.Stat(workingCopyPath); !os.IsNotExist(err) {
			c.Printf(colors.Yellow("Not cloning %s as a directory already exists at %s\n"), repo.FullRepoName, workingCopyPath)
			continue
		}

		c.Printf("Forking and cloning %s into %s/%s\n", repo.FullRepoName, parentPath, repo.RepoName)
		err = exec.Execute(c, parentPath, "gh", "repo", "fork", "--clone=true", repo.FullRepoName)
		if err != nil {
			c.Printf(colors.Red("Error when cloning %s: %s\n"), repo.FullRepoName, err)
			continue
		}

		c.Printf("Creating branch %s in %s/%s\n", dir.Name, parentPath, repo.RepoName)
		err = exec.Execute(c, workingCopyPath, "git", "checkout", "-b", dir.Name)
		if err != nil {
			c.Printf(colors.Red("Error when creating branch: %s\n"), err)
			continue
		}
	}
}
