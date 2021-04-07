package clone

import (
	"github.com/spf13/cobra"
)

func CreateCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: " ", // TODO
		Run:   run,
	}

	return cmd
}

func run(*cobra.Command, []string) {

}
