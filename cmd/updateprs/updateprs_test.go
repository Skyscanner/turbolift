package updateprs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/prompt"
	"github.com/skyscanner/turbolift/internal/testsupport"
)

func TestValidateFlagsNoneSet(t *testing.T) {
	err := validateFlags(false, false, false)
	assert.Error(t, err)
	assert.Equal(t, "update-prs needs one and only one action flag", err.Error())
}

func TestValidateFlagsMultipleSet(t *testing.T) {
	err := validateFlags(true, false, true)
	assert.Error(t, err)
	assert.Equal(t, "update-prs needs one and only one action flag", err.Error())
}

func TestValidateFlagsSingleSet(t *testing.T) {
	err := validateFlags(true, false, false)
	assert.NoError(t, err)
}

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
		{"close_pull_request", "work/org/repo1", filepath.Base(tempDir)},
		{"close_pull_request", "work/org/repo2", filepath.Base(tempDir)},
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
		{"close_pull_request", "work/org/repo1", filepath.Base(tempDir)},
		{"close_pull_request", "work/org/repo2", filepath.Base(tempDir)},
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
		{"close_pull_request", "work/org/repo1", filepath.Base(tempDir)},
		{"close_pull_request", "work/org/repo2", filepath.Base(tempDir)},
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
		{"update_pr_description", "work/org/repo1", "PR title", "PR body"},
		{"update_pr_description", "work/org/repo2", "PR title", "PR body"},
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
		{"update_pr_description", "work/org/repo1", "Updated PR title", "Updated PR body"},
		{"update_pr_description", "work/org/repo2", "Updated PR title", "Updated PR body"},
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
		{"update_pr_description", "work/org/repo1", "custom PR title", "custom PR body"},
		{"update_pr_description", "work/org/repo2", "custom PR title", "custom PR body"},
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

func TestItPushesNewCommits(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	tempDir := testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runPushCommandAuto()
	assert.NoError(t, err)
	assert.Contains(t, out, "Pushing changes in org/repo1 to origin")
	assert.Contains(t, out, "Pushing changes in org/repo2 to origin")
	assert.Contains(t, out, "turbolift update-prs completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeGit.AssertCalledWith(t, [][]string{
		{"push", "work/org/repo1", filepath.Base(tempDir)},
		{"push", "work/org/repo2", filepath.Base(tempDir)},
	})
}

func TestItLogsPushErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	tempDir := testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runPushCommandAuto()
	assert.NoError(t, err)
	assert.Contains(t, out, "Pushing changes in org/repo1 to origin")
	assert.Contains(t, out, "Pushing changes in org/repo2 to origin")
	assert.Contains(t, out, "turbolift update-prs completed with errors")
	assert.Contains(t, out, "2 errored")

	fakeGit.AssertCalledWith(t, [][]string{
		{"push", "work/org/repo1", filepath.Base(tempDir)},
		{"push", "work/org/repo2", filepath.Base(tempDir)},
	})
}

func TestItPushesPerRepoBranchWhenAnnotated(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	tempDir := testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	// Rewrite repos.txt after setup so the annotation doesn't get mangled
	// into directory names during PrepareTempCampaign's dir creation.
	if err := os.WriteFile("repos.txt", []byte("org/repo1 # branch=pr-branch\norg/repo2\n"), 0o644); err != nil {
		t.Fatalf("write repos.txt: %v", err)
	}

	_, err := runPushCommandAuto()
	assert.NoError(t, err)

	campaignBranch := filepath.Base(tempDir)
	fakeGit.AssertCalledWith(t, [][]string{
		{"push", "work/org/repo1", "pr-branch"},
		{"push", "work/org/repo2", campaignBranch},
	})
}

func TestItClosesPerRepoBranchWhenAnnotated(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	tempDir := testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	if err := os.WriteFile("repos.txt", []byte("org/repo1 # branch=pr-branch\norg/repo2\n"), 0o644); err != nil {
		t.Fatalf("write repos.txt: %v", err)
	}

	_, err := runCloseCommandAuto()
	assert.NoError(t, err)

	campaignBranch := filepath.Base(tempDir)
	fakeGitHub.AssertCalledWith(t, [][]string{
		{"close_pull_request", "work/org/repo1", "pr-branch"},
		{"close_pull_request", "work/org/repo2", campaignBranch},
	})
}

func TestItDoesNotPushIfNotConfirmed(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakePrompt := prompt.NewFakePromptNo()
	p = fakePrompt

	_ = testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runPushCommandConfirm()
	assert.NoError(t, err)
	assert.NotContains(t, out, "Pushing changes in org/repo1 to origin")
	assert.NotContains(t, out, "Pushing changes in org/repo2 to origin")
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

func runPushCommandAuto() (string, error) {
	cmd := NewUpdatePRsCmd()
	pushFlag = true
	yesFlag = true
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runPushCommandConfirm() (string, error) {
	cmd := NewUpdatePRsCmd()
	pushFlag = true
	yesFlag = false
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
