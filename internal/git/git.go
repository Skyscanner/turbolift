package git

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type Git interface {
	Checkout(output io.Writer, workingDir string, branch string) error
	Push(stdout io.Writer, workingDir string, remote string, branchName string) error
}

type RealGit struct {
}

func (r *RealGit) Checkout(output io.Writer, workingDir string, branchName string) error {
	return execInstance.Execute(output, workingDir, "git", "checkout", "-b", branchName)
}

func (r *RealGit) Push(output io.Writer, workingDir string, remote string, branchName string) error {
	return execInstance.Execute(output, workingDir, "git", "push", "-u", remote, branchName)
}

func NewRealGit() *RealGit {
	return &RealGit{}
}
