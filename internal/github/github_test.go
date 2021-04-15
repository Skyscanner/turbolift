package github

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/stretchr/testify/assert"
	"strings"
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
	sb := strings.Builder{}
	err := NewRealGitHub().ForkAndClone(&sb, "work/org", "org/repo1")

	if err != nil {
		return sb.String(), err
	}
	return sb.String(), nil
}
