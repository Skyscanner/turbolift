package git

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
		{"work/org/repo1", "git", "checkout", "-b", "some_branch"},
	})
}

func TestItReturnsNilErrorOnSuccess(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "git", "checkout", "-b", "some_branch"},
	})
}

func runAndCaptureOutput() (string, error) {
	sb := strings.Builder{}
	err := NewRealGit().Checkout(&sb, "work/org/repo1", "some_branch")

	if err != nil {
		return sb.String(), err
	}
	return sb.String(), nil
}
