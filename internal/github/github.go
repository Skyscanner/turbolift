package github

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type PullRequest struct {
	title string
	body  string
}

type GitHub interface {
	ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error
	CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error)
}

type RealGitHub struct {
}

func (r *RealGitHub) CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error) {
	panic("implement me")
}

func (r *RealGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	return execInstance.Execute(output, workingDir, "gh", "repo", "fork", "--clone=true", fullRepoName)
}

func NewRealGitHub() *RealGitHub {
	return &RealGitHub{}
}
