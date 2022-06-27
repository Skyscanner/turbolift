package update_prs

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/prompt"
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestItLogsClosePrErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub

	tempDir := testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCloseCommandAuto()
	assert.NoError(t, err)
	assert.Contains(t, out, "Closing PR in org/repo1")
	assert.Contains(t, out, "Closing PR in org/repo2")
	assert.Contains(t, out, "turbolift update-prs completed with errors")
	assert.Contains(t, out, "2 errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", filepath.Base(tempDir)},
		{"work/org/repo2", filepath.Base(tempDir)},
	})
}

func TestItClosesPrsSuccessfully(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub

	tempDir := testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCloseCommandAuto()
	assert.NoError(t, err)
	assert.Contains(t, out, "Closing PR in org/repo1")
	assert.Contains(t, out, "Closing PR in org/repo2")
	assert.Contains(t, out, "turbolift update-prs completed")
	assert.Contains(t, out, "2 OK")
	assert.NotContains(t, out, "error")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", filepath.Base(tempDir)},
		{"work/org/repo2", filepath.Base(tempDir)},
	})
}

func TestNoPRFound(t *testing.T) {
	fakeGitHub := github.NewAlwaysThrowNoPRFound()
	gh = fakeGitHub

	tempDir := testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCloseCommandAuto()
	assert.NoError(t, err)
	assert.Contains(t, out, "no PR found for work/org/repo1 and branch "+filepath.Base(tempDir))
	assert.Contains(t, out, "no PR found for work/org/repo2 and branch "+filepath.Base(tempDir))
	assert.Contains(t, out, "turbolift update-prs completed")
	assert.Contains(t, out, "0 OK, 2 skipped")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", filepath.Base(tempDir)},
		{"work/org/repo2", filepath.Base(tempDir)},
	})
}

func TestItDoesNotClosePRsIfNotConfirmed(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakePrompt := prompt.NewFakePromptNo()
	p = fakePrompt

	_ = testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCloseCommandConfirm()
	assert.NoError(t, err)
	assert.NotContains(t, out, "Closing PR in org/repo1")
	assert.NotContains(t, out, "Closing PR in org/repo2")
	assert.NotContains(t, out, "turbolift update-prs completed")
	assert.NotContains(t, out, "2 OK")

	fakeGitHub.AssertCalledWith(t, [][]string{})
}

func runCloseCommandAuto() (string, error) {
	cmd := NewUpdatePRsCmd()
	closeFlag = true
	yesFlag = true
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runCloseCommandConfirm() (string, error) {
	cmd := NewUpdatePRsCmd()
	closeFlag = true
	yesFlag = false
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
