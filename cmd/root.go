package cmd

import (
	cloneCmd "github.com/skyscanner/turbolift/cmd/clone"
	createPrsCmd "github.com/skyscanner/turbolift/cmd/create_prs"
	commitCmd "github.com/skyscanner/turbolift/cmd/commit"
	foreachCmd "github.com/skyscanner/turbolift/cmd/foreach"
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
	rootCmd.AddCommand(commitCmd.NewCommitCmd())
	rootCmd.AddCommand(cloneCmd.NewCloneCmd())
	rootCmd.AddCommand(initCmd.NewInitCmd())
	rootCmd.AddCommand(foreachCmd.NewForeachCmd())
	rootCmd.AddCommand(createPrsCmd.NewCreatePRsCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
