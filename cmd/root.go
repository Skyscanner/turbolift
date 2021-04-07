package cmd

import (
	cloneCmd "github.com/skyscanner/turbolift/cmd/clone"
	initCmd "github.com/skyscanner/turbolift/cmd/init"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "turbolift",
	Short: "Turbolift",
	Long:  `Mass refactoring tool for repositories in GitHub`,
}

func init() {
	rootCmd.AddCommand(cloneCmd.CreateCloneCmd())
	rootCmd.AddCommand(initCmd.CreateInitCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
