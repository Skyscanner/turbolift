package executor

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type FakeExecutor struct {
	Handler          func(workingDir string, name string, args ...string) error
	ReturningHandler func(workingDir string, name string, args ...string) (string, error)
	calls            [][]string
}

func (e *FakeExecutor) Execute(_ io.Writer, workingDir string, name string, args ...string) error {
	allArgs := append([]string{workingDir, name}, args...)
	e.calls = append(e.calls, allArgs)
	return e.Handler(workingDir, name, args...)
}

func (e *FakeExecutor) ExecuteAndCapture(_ io.Writer, workingDir string, name string, args ...string) (string, error) {
	allArgs := append([]string{workingDir, name}, args...)
	e.calls = append(e.calls, allArgs)
	return e.ReturningHandler(workingDir, name, args...)
}

func (e *FakeExecutor) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, e.calls)
}

func NewFakeExecutor(handler func(string, string, ...string) error, returningHandler func(string, string, ...string) (string, error)) *FakeExecutor {
	return &FakeExecutor{
		Handler:          handler,
		ReturningHandler: returningHandler,
		calls:            [][]string{},
	}
}

func NewAlwaysSucceedsFakeExecutor() *FakeExecutor {
	return NewFakeExecutor(func(s string, s2 string, s3 ...string) error {
		return nil
	}, func(s string, s2 string, s3 ...string) (string, error) {
		return "", nil
	})
}

func NewAlwaysFailsFakeExecutor() *FakeExecutor {
	return NewFakeExecutor(func(s string, s2 string, s3 ...string) error {
		return errors.New("synthetic error")
	}, func(s string, s2 string, s3 ...string) (string, error) {
		return "", errors.New("synthetic error")
	})
}
