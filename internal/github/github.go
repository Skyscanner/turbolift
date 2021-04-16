package github

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type GitHub interface {
	ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error
}

type RealGitHub struct {
}

func (r *RealGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	return execInstance.Execute(output, workingDir, "gh", "repo", "fork", "--clone=true", fullRepoName)
}

func NewRealGitHub() *RealGitHub {
	return &RealGitHub{}
}
