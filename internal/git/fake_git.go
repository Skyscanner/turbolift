package git

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

type FakeGit struct {
	handler func(c *cobra.Command, workingDir string, branchName string) error
	calls   [][]string
}

func (f *FakeGit) Checkout(c *cobra.Command, workingDir string, branch string) error {
	f.calls = append(f.calls, []string{workingDir, branch})
	return f.handler(c, workingDir, branch)
}

func (f *FakeGit) ForkAndClone(c *cobra.Command, workingDir string, branchName string) error {
	f.calls = append(f.calls, []string{workingDir, branchName})
	return f.handler(c, workingDir, branchName)
}

func (f *FakeGit) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGit(h func(c *cobra.Command, workingDir string, branchName string) error) *FakeGit {
	return &FakeGit{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGit() *FakeGit {
	return NewFakeGit(func(c *cobra.Command, workingDir string, branchName string) error {
		return nil
	})
}

func NewAlwaysFailsFakeGit() *FakeGit {
	return NewFakeGit(func(c *cobra.Command, workingDir string, branchName string) error {
		return errors.New("synthetic error")
	})
}
