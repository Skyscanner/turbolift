package git

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io"
	"os"
	"strconv"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type Git interface {
	Checkout(output io.Writer, workingDir string, branch string) error
	Push(stdout io.Writer, workingDir string, remote string, branchName string) error
	Commit(output io.Writer, workingDir string, message string) error
	IsRepoChanged(output io.Writer, workingDir string) (bool, error)
}

type RealGit struct {
}

func (r *RealGit) Checkout(output io.Writer, workingDir string, branchName string) error {
	return execInstance.Execute(output, workingDir, "git", "checkout", "-b", branchName)
}

func (r *RealGit) Push(output io.Writer, workingDir string, remote string, branchName string) error {
	return execInstance.Execute(output, workingDir, "git", "push", "-u", remote, branchName)
}

func (r *RealGit) Commit(output io.Writer, workingDir string, message string) error {
	return execInstance.Execute(output, workingDir, "git", "commit", "--all", "--message", message)
}

func (r *RealGit) IsRepoChanged(output io.Writer, workingDir string) (bool, error) {
	shellCommand := os.Getenv("SHELL")
	if shellCommand == "" {
		shellCommand = "sh"
	}
	shellArgs := []string{"-c", "git status --porcelain=v1 | wc -l | tr -d '[:space:]'"}
	commandOutput, err := execInstance.ExecuteAndCapture(output, workingDir, shellCommand, shellArgs...)

	if err != nil {
		return false, err
	}

	diffSize, err := strconv.Atoi(commandOutput)
	if err != nil {
		return false, err
	}

	return diffSize > 0, nil
}

func NewRealGit() *RealGit {
	return &RealGit{}
}
