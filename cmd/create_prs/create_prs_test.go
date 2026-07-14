/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * https://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package create_prs

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/prompt"
	"github.com/skyscanner/turbolift/internal/testsupport"
)

func TestItWarnsIfDescriptionFileTemplateIsUnchanged(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit
	fakePrompt := prompt.NewFakePromptNo()
	p = fakePrompt

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	testsupport.UseDefaultPrDescription()

	out, err := runCommand()
	assert.NoError(t, err)
	assert.NotContains(t, out, "Creating PR in org/repo1")
	assert.NotContains(t, out, "Creating PR in org/repo2")
	assert.NotContains(t, out, "turbolift create-prs completed")
	assert.NotContains(t, out, "2 OK, 0 skipped")

	fakePrompt.AssertCalledWith(t, "It looks like the PR title and/or description may not have been updated in README.md. Are you sure you want to proceed?")
}

func TestItWarnsIfOnlyPrTitleIsUnchanged(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit
	fakePrompt := prompt.NewFakePromptNo()
	p = fakePrompt

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	testsupport.UsePrTitleTodoOnly()

	out, err := runCommand()
	assert.NoError(t, err)
	assert.NotContains(t, out, "Creating PR in org/repo1")
	assert.NotContains(t, out, "Creating PR in org/repo2")
	assert.NotContains(t, out, "turbolift create-prs completed")
	assert.NotContains(t, out, "2 OK, 0 skipped")

	fakePrompt.AssertCalledWith(t, "It looks like the PR title and/or description may not have been updated in README.md. Are you sure you want to proceed?")
}

func TestItWarnsIfOnlyPrBodyIsUnchanged(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit
	fakePrompt := prompt.NewFakePromptNo()
	p = fakePrompt

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	testsupport.UsePrBodyTodoOnly()

	out, err := runCommand()
	assert.NoError(t, err)
	assert.NotContains(t, out, "Creating PR in org/repo1")
	assert.NotContains(t, out, "Creating PR in org/repo2")
	assert.NotContains(t, out, "turbolift create-prs completed")
	assert.NotContains(t, out, "2 OK, 0 skipped")

	fakePrompt.AssertCalledWith(t, "It looks like the PR title and/or description may not have been updated in README.md. Are you sure you want to proceed?")
}

func TestItWarnsIfDescriptionFileIsEmpty(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit
	fakePrompt := prompt.NewFakePromptNo()
	p = fakePrompt

	customDescriptionFileName := "custom.md"

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	testsupport.CreateOrUpdatePrDescriptionFile(customDescriptionFileName, "", "")

	out, err := runCommandWithAlternativeDescriptionFile(customDescriptionFileName)
	assert.NoError(t, err)
	assert.NotContains(t, out, "Creating PR in org/repo1")
	assert.NotContains(t, out, "Creating PR in org/repo2")
	assert.NotContains(t, out, "turbolift create-prs completed")
	assert.NotContains(t, out, "2 OK, 0 skipped")

	fakePrompt.AssertCalledWith(t, "It looks like the PR title and/or description may not have been updated in custom.md. Are you sure you want to proceed?")
}

func TestItLogsCreatePrErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Creating PR in org/repo1")
	assert.Contains(t, out, "Creating PR in org/repo2")
	assert.Contains(t, out, "turbolift create-prs completed with errors")
	assert.Contains(t, out, "2 errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"create_pull_request", "work/org/repo1", "PR title"},
		{"create_pull_request", "work/org/repo2", "PR title"},
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
		{"create_pull_request", "work/org/repo1", "PR title"},
		{"create_pull_request", "work/org/repo2", "PR title"},
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
		{"create_pull_request", "work/org/repo1", "PR title"},
		{"create_pull_request", "work/org/repo2", "PR title"},
	})
}

func TestItLogsCreateDraftPr(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommandDraft()
	assert.NoError(t, err)
	assert.Contains(t, out, "Creating Draft PR in org/repo1")
	assert.Contains(t, out, "Creating Draft PR in org/repo2")
	assert.Contains(t, out, "turbolift create-prs completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"create_pull_request", "work/org/repo1", "PR title"},
		{"create_pull_request", "work/org/repo2", "PR title"},
	})
}

func TestItCreatesPrsFromAlternativeDescriptionFile(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	customDescriptionFileName := "custom.md"

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	testsupport.CreateOrUpdatePrDescriptionFile(customDescriptionFileName, "custom PR title", "custom PR body")

	out, err := runCommandWithAlternativeDescriptionFile(customDescriptionFileName)
	assert.NoError(t, err)
	assert.Contains(t, out, "Reading campaign data (repos.txt, custom.md)")
	assert.Contains(t, out, "Creating PR in org/repo1")
	assert.Contains(t, out, "Creating PR in org/repo2")
	assert.Contains(t, out, "turbolift create-prs completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"create_pull_request", "work/org/repo1", "custom PR title"},
		{"create_pull_request", "work/org/repo2", "custom PR title"},
	})
}

func TestItPushesPerRepoBranchWhenAnnotated(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	// repo1 has an annotated branch; repo2 falls back to the campaign name.
	// We rewrite repos.txt after PrepareTempCampaign so the annotation
	// doesn't get mangled into directory names during setup.
	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	reposContent := "org/repo1 # branch=pr-branch\norg/repo2\n"
	if err := os.WriteFile("repos.txt", []byte(reposContent), 0o644); err != nil {
		t.Fatalf("write repos.txt: %v", err)
	}

	_, err := runCommand()
	assert.NoError(t, err)

	// Confirm push was called per-repo with the correct branch. The campaign
	// name is the basename of the temp directory (starts with "turbolift-").
	campaignBranch := testsupport.Pwd()
	fakeGit.AssertCalledWith(t, [][]string{
		{"push", "work/org/repo1", "pr-branch"},
		{"push", "work/org/repo2", campaignBranch},
	})
}

func runCommand() (string, error) {
	cmd := NewCreatePRsCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	return outBuffer.String(), err
}

func runCommandWithAlternativeDescriptionFile(fileName string) (string, error) {
	cmd := NewCreatePRsCmd()
	prDescriptionFile = fileName
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	return outBuffer.String(), err
}

func runCommandDraft() (string, error) {
	cmd := NewCreatePRsCmd()
	isDraft = true
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	return outBuffer.String(), err
}
