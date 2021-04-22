package create_prs

import (
	"bytes"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestItLogsCreatePrErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "❌ Creating PR in org/repo1")
	assert.Contains(t, out, "❌ Creating PR in org/repo2")
	assert.Contains(t, out, "turbolift create-prs completed with errors")
	assert.Contains(t, out, "2 errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "PR title"},
		{"work/org/repo2", "PR title"},
	})
}

func TestItLogsCreatePrSkippedButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysReturnsFalseFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "No PR created in org/repo1")
	assert.Contains(t, out, "No PR created in org/repo2")
	assert.Contains(t, out, "turbolift create-prs completed")
	assert.Contains(t, out, "0 OK, 2 skipped")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "PR title"},
		{"work/org/repo2", "PR title"},
	})
}

func TestItLogsCreatePrsSucceeds(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift create-prs completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "PR title"},
		{"work/org/repo2", "PR title"},
	})
}

func runCommand() (string, error) {
	cmd := NewCreatePRsCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()

	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
