package github

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type PullRequest struct {
	Title        string
	Body         string
	UpstreamRepo string
}

type GitHub interface {
	ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error
	CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error)
}

type RealGitHub struct {
}

func (r *RealGitHub) CreatePullRequest(output io.Writer, workingDir string, pr PullRequest) (didCreate bool, err error) {
	err = execInstance.Execute(output, workingDir, "gh", "pr", "create", "--title", pr.Title, "--body", pr.Body, "--repo", pr.UpstreamRepo)
	return err != nil, err
}

func (r *RealGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	return execInstance.Execute(output, workingDir, "gh", "repo", "fork", "--clone=true", fullRepoName)
}

func NewRealGitHub() *RealGitHub {
	return &RealGitHub{}
}
