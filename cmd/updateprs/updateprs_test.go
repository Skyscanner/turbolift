package updateprs

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/prompt"
	"github.com/skyscanner/turbolift/internal/testsupport"
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

func TestItLogsUpdateDescriptionErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runUpdateDescriptionCommandAuto("README.md")
	assert.NoError(t, err)
	assert.Contains(t, out, "Updating PR description in org/repo1")
	assert.Contains(t, out, "Updating PR description in org/repo2")
	assert.Contains(t, out, "turbolift update-prs completed with errors")
	assert.Contains(t, out, "2 errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "PR title", "PR body"},
		{"work/org/repo2", "PR title", "PR body"},
	})
}

func TestItUpdatesDescriptionsSuccessfully(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	testsupport.CreateOrUpdatePrDescriptionFile("README.md", "Updated PR title", "Updated PR body")

	out, err := runUpdateDescriptionCommandAuto("README.md")
	assert.NoError(t, err)
	assert.Contains(t, out, "Updating PR description in org/repo1")
	assert.Contains(t, out, "Updating PR description in org/repo2")
	assert.Contains(t, out, "turbolift update-prs completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "Updated PR title", "Updated PR body"},
		{"work/org/repo2", "Updated PR title", "Updated PR body"},
	})
}

func TestItUpdatesDescriptionsFromAlternativeFile(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	testsupport.CreateOrUpdatePrDescriptionFile("custom.md", "custom PR title", "custom PR body")

	out, err := runUpdateDescriptionCommandAuto("custom.md")
	assert.NoError(t, err)
	assert.Contains(t, out, "Updating PR description in org/repo1")
	assert.Contains(t, out, "Updating PR description in org/repo2")
	assert.Contains(t, out, "turbolift update-prs completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "custom PR title", "custom PR body"},
		{"work/org/repo2", "custom PR title", "custom PR body"},
	})
}

func TestItDoesNotUpdateDescriptionsIfNotConfirmed(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakePrompt := prompt.NewFakePromptNo()
	p = fakePrompt

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runUpdateDescriptionCommandConfirm()
	assert.NoError(t, err)
	assert.NotContains(t, out, "Updating PR description in org/repo1")
	assert.NotContains(t, out, "Updating PR description in org/repo2")
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

func runUpdateDescriptionCommandAuto(descriptionFile string) (string, error) {
	cmd := NewUpdatePRsCmd()
	updateDescriptionFlag = true
	yesFlag = true
	prDescriptionFile = descriptionFile
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runUpdateDescriptionCommandConfirm() (string, error) {
	cmd := NewUpdatePRsCmd()
	updateDescriptionFlag = true
	yesFlag = false
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
