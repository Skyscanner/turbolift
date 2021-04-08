package clone

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/spf13/cobra"
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

func run(c *cobra.Command, args []string) {
	err := exec.Execute("gh", "repo", "clone", "mshell/mshell-tools")
	if err != nil {
		panic(err)
	}
}
