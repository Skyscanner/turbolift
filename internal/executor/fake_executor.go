package executor

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type FakeExecutor struct {
	Handler func(workingDir string, name string, args ...string) error
	calls   [][]string
}

func (e *FakeExecutor) Execute(workingDir string, name string, args ...string) error {
	allArgs := append([]string{workingDir, name}, args...)
	e.calls = append(e.calls, allArgs)
	return e.Handler(workingDir, name, args...)
}

func (e *FakeExecutor) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, e.calls)
}

func NewFakeExecutor(h func(string, string, ...string) error) *FakeExecutor {
	return &FakeExecutor{
		Handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeExecutor() *FakeExecutor {
	return NewFakeExecutor(func(s string, s2 string, s3 ...string) error {
		return nil
	})
}

func NewAlwaysFailsFakeExecutor() *FakeExecutor {
	return NewFakeExecutor(func(s string, s2 string, s3 ...string) error {
		return errors.New("synthetic error")
	})
}
