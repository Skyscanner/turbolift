package git

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type FakeGit struct {
	handler func(output io.Writer, workingDir string, branchName string) error
	calls   [][]string
}

func (f *FakeGit) Checkout(output io.Writer, workingDir string, branch string) error {
	f.calls = append(f.calls, []string{workingDir, branch})
	return f.handler(output, workingDir, branch)
}

func (f *FakeGit) ForkAndClone(output io.Writer, workingDir string, branchName string) error {
	f.calls = append(f.calls, []string{workingDir, branchName})
	return f.handler(output, workingDir, branchName)
}

func (f *FakeGit) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGit(h func(output io.Writer, workingDir string, branchName string) error) *FakeGit {
	return &FakeGit{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGit() *FakeGit {
	return NewFakeGit(func(output io.Writer, workingDir string, branchName string) error {
		return nil
	})
}

func NewAlwaysFailsFakeGit() *FakeGit {
	return NewFakeGit(func(output io.Writer, workingDir string, branchName string) error {
		return errors.New("synthetic error")
	})
}
