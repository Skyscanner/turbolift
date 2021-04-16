package github

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type FakeGitHub struct {
	handler func(output io.Writer, workingDir string, fullRepoName string) error
	calls   [][]string
}

func (f *FakeGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	f.calls = append(f.calls, []string{workingDir, fullRepoName})
	return f.handler(output, workingDir, fullRepoName)
}

func (f *FakeGitHub) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGitHub(h func(output io.Writer, workingDir string, fullRepoName string) error) *FakeGitHub {
	return &FakeGitHub{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) error {
		return nil
	})
}

func NewAlwaysFailsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) error {
		return errors.New("synthetic error")
	})
}
