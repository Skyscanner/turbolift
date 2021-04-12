package git

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/spf13/cobra"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type Git interface {
	Checkout(c *cobra.Command, workingDir string, branch string) error
}

type RealGit struct {
}

func (r *RealGit) Checkout(c *cobra.Command, workingDir string, branchName string) error {
	return execInstance.Execute(c, workingDir, "git", "checkout", "-b", branchName)
}

func NewRealGit() *RealGit {
	return &RealGit{}
}
