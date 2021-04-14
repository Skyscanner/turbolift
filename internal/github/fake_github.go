package github

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type FakeGitHub struct {
	handler func(output io.Writer, workingDir string, fullRepoName string) (bool, error)
	calls   [][]string
}

func (f *FakeGitHub) CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error) {
	f.calls = append(f.calls, []string{workingDir, metadata.title})
	return f.handler(output, workingDir, "")
}

func (f *FakeGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	f.calls = append(f.calls, []string{workingDir, fullRepoName})
	_, err := f.handler(output, workingDir, fullRepoName)
	return err
}

func (f *FakeGitHub) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGitHub(h func(output io.Writer, workingDir string, fullRepoName string) (bool, error)) *FakeGitHub {
	return &FakeGitHub{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) (bool, error) {
		return true, nil
	})
}

func NewAlwaysFailsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) (bool, error) {
		return false, errors.New("synthetic error")
	})
}

func NewAlwaysReturnsFalseFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) (bool, error) {
		return false, nil
	})
}
