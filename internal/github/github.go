package github

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/spf13/cobra"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type GitHub interface {
	ForkAndClone(c *cobra.Command, workingDir string, fullRepoName string) error
}

type RealGitHub struct {
}

func (r *RealGitHub) ForkAndClone(c *cobra.Command, workingDir string, fullRepoName string) error {
	return execInstance.Execute(c, workingDir, "gh", "repo", "fork", "--clone=true", fullRepoName)
}

func NewRealGitHub() *RealGitHub {
	return &RealGitHub{}
}
