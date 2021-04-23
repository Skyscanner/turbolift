package git

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type FakeGit struct {
	handler func(output io.Writer, call []string) (bool, error)
	calls   [][]string
}

func (f *FakeGit) Checkout(output io.Writer, workingDir string, branch string) error {
	call := []string{"checkout", workingDir, branch}
	f.calls = append(f.calls, call)
	_, err := f.handler(output, call)
	return err
}

func (f *FakeGit) Commit(output io.Writer, workingDir string, message string) error {
	call := []string{"commit", workingDir, message}
	f.calls = append(f.calls, call)
	_, err := f.handler(output, call)
	return err
}

func (f *FakeGit) IsRepoChanged(output io.Writer, workingDir string) (bool, error) {
	call := []string{"isRepoChanged", workingDir}
	f.calls = append(f.calls, call)
	result, err := f.handler(output, call)
	return result, err
}

func (f *FakeGit) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGit(h func(io.Writer, []string) (bool, error)) *FakeGit {
	return &FakeGit{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGit() *FakeGit {
	return NewFakeGit(func(io.Writer, []string) (bool, error) {
		return true, nil
	})
}

func NewAlwaysFailsFakeGit() *FakeGit {
	return NewFakeGit(func(io.Writer, []string) (bool, error) {
		return false, errors.New("synthetic error")
	})
}
