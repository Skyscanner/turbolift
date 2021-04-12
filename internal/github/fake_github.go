package github

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

type FakeGitHub struct {
	handler func(c *cobra.Command, workingDir string, fullRepoName string) error
	calls   [][]string
}

func (f *FakeGitHub) ForkAndClone(c *cobra.Command, workingDir string, fullRepoName string) error {
	f.calls = append(f.calls, []string{workingDir, fullRepoName})
	return f.handler(c, workingDir, fullRepoName)
}

func (f *FakeGitHub) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGitHub(h func(c *cobra.Command, workingDir string, fullRepoName string) error) *FakeGitHub {
	return &FakeGitHub{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(c *cobra.Command, workingDir string, fullRepoName string) error {
		return nil
	})
}

func NewAlwaysFailsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(c *cobra.Command, workingDir string, fullRepoName string) error {
		return errors.New("synthetic error")
	})
}
