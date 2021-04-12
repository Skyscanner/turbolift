package github

import (
	"bytes"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestItReturnsErrorOnFailure(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
	})
}

func TestItReturnsNilErrorOnSuccess(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
	})
}

func runAndCaptureOutput() (string, error) {
	c := &cobra.Command{}
	outBuffer := bytes.NewBufferString("")
	c.SetOut(outBuffer)
	err := NewRealGitHub().ForkAndClone(c, "work/org", "org/repo1")

	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
