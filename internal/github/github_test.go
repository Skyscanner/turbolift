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

package github

import (
	"errors"
	"github.com/skyscanner/turbolift/internal/git"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/executor"
)

func TestItReturnsErrorOnFailedFork(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runForkAndCloneAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
	})
}

func TestItReturnsNilErrorOnSuccessfulFork(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runForkAndCloneAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
	})
}

func TestItReturnsErrorOnFailedClone(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runCloneAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "clone", "org/repo1"},
	})
}

func TestItReturnsNilErrorOnSuccessfulClone(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runCloneAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "clone", "org/repo1"},
	})
}

func TestItReturnsErrorOnFailedCreatePr(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrAndCaptureOutput()
	assert.Error(t, err)
	assert.False(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1"},
	})
}

func TestItReturnsFalseAndNilErrorOnNoOpCreatePr(t *testing.T) {
	fakeExecutor := executor.NewFakeExecutor(func(workingDir string, name string, args ...string) error {
		return nil
	}, func(workingDir string, name string, args ...string) (string, error) {
		return "... GraphQL error: No commits between A and B ...", errors.New("synthetic error")
	})
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrAndCaptureOutput()
	assert.NoError(t, err)
	assert.False(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1"},
	})
}

func TestItSuccessfulCreatesADraftPr(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreateDraftPrAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1", "--draft"},
	})
}

func TestItReturnsTrueAndNilErrorOnSuccessfulCreatePr(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1"},
	})
}

func TestItReturnsErrorOnFailedGetDefaultBranchName(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, _, err := runGetDefaultBranchNameAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org1/repo1", "gh", "repo", "view", "org1/repo1", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name"},
	})
}

func TestItReturnsNilErrorOnSuccessfulGetDefaultBranchName(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, _, err := runGetDefaultBranchNameAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org1/repo1", "gh", "repo", "view", "org1/repo1", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name"},
	})
}

func TestItReturnsErrorOnFailedUpdatePrDescription(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runUpdatePrDescriptionAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "pr", "edit", "--title", "new title", "--body", "new body"},
	})
}

func TestItReturnsNilErrorOnSuccessfulUpdatePrDescription(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runUpdatePrDescriptionAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "pr", "edit", "--title", "new title", "--body", "new body"},
	})
}

func TestItReturnsTrueAndNilErrorWhenRepoIsFork(t *testing.T) {
	fakeExecutor := newIsForkFakeExecutor("true", nil)
	execInstance = fakeExecutor

	isFork, _, err := runIsForkAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, isFork)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "repo", "view", "https://github.com/dummyOrg/dummyRepo.git", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
		{"work/org/repo1", "gh", "repo", "view", "org/repo1", "--json", "isFork"},
	})
}

func TestItReturnsFalseAndNilErrorWhenRepoIsNotFork(t *testing.T) {
	fakeExecutor := newIsForkFakeExecutor("false", nil)
	execInstance = fakeExecutor

	isFork, _, err := runIsForkAndCaptureOutput()
	assert.NoError(t, err)
	assert.False(t, isFork)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "repo", "view", "https://github.com/dummyOrg/dummyRepo.git", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
		{"work/org/repo1", "gh", "repo", "view", "org/repo1", "--json", "isFork"},
	})
}

func TestItReturnsErrorOnFailedIsFork(t *testing.T) {
	fakeExecutor := newIsForkFakeExecutor("", errors.New("synthetic error"))
	execInstance = fakeExecutor

	_, _, err := runIsForkAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "repo", "view", "https://github.com/dummyOrg/dummyRepo.git", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
		{"work/org/repo1", "gh", "repo", "view", "org/repo1", "--json", "isFork"},
	})
}

func TestItReturnsTrueAndNilErrorWhenUserHasOpenUpstreamPRs(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsAndReturnsTrueFakeExecutor()
	execInstance = fakeExecutor

	hasOpenPRs, _, err := runUserHasOpenUpstreamPRsAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, hasOpenPRs)

	currentDir, _ := os.Getwd()

	fakeExecutor.AssertCalledWith(t, [][]string{
		{currentDir, "gh", "pr", "list", "--repo", "org/repo1", "--author", "@me", "--state", "open", "--limit", "1", "--json", "number", "--jq", "length > 0"},
	})
}

func TestItReturnsFalseAndNilErrorWhenUserHasNoOpenUpstreamPRs(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsAndReturnsFalseFakeExecutor()
	execInstance = fakeExecutor

	hasOpenPRs, _, err := runUserHasOpenUpstreamPRsAndCaptureOutput()
	assert.NoError(t, err)
	assert.False(t, hasOpenPRs)

	currentDir, _ := os.Getwd()

	fakeExecutor.AssertCalledWith(t, [][]string{
		{currentDir, "gh", "pr", "list", "--repo", "org/repo1", "--author", "@me", "--state", "open", "--limit", "1", "--json", "number", "--jq", "length > 0"},
	})
}

func TestItReturnsErrorOnFailedUserHasOpenUpstreamPRs(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, _, err := runUserHasOpenUpstreamPRsAndCaptureOutput()
	assert.Error(t, err)

	currentDir, _ := os.Getwd()

	fakeExecutor.AssertCalledWith(t, [][]string{
		{currentDir, "gh", "pr", "list", "--repo", "org/repo1", "--author", "@me", "--state", "open", "--limit", "1", "--json", "number", "--jq", "length > 0"},
	})
}

func TestItReturnsErrorOnFailedGetOriginRepoName(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, _, err := runGetOriginRepoNameAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "repo", "view", "https://github.com/dummyOrg/dummyRepo.git", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
	})
}

func TestItReturnsNilErrorOnSuccessfulGetOriginRepoName(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, _, err := runGetOriginRepoNameAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "repo", "view", "https://github.com/dummyOrg/dummyRepo.git", "--json", "nameWithOwner", "--jq", ".nameWithOwner"},
	})
}

func runForkAndCloneAndCaptureOutput() (string, error) {
	sb := strings.Builder{}
	err := NewRealGitHub().ForkAndClone(&sb, "work/org", "org/repo1")

	return sb.String(), err
}

func runCloneAndCaptureOutput() (string, error) {
	sb := strings.Builder{}
	err := NewRealGitHub().Clone(&sb, "work/org", "org/repo1")

	return sb.String(), err
}

func runCreatePrAndCaptureOutput() (bool, string, error) {
	sb := strings.Builder{}
	didCreatePr, err := NewRealGitHub().CreatePullRequest(&sb, "work/org/repo1", PullRequest{
		Title:        "some title",
		Body:         "some body",
		UpstreamRepo: "org/repo1",
	})

	return didCreatePr, sb.String(), err
}

func runCreateDraftPrAndCaptureOutput() (bool, string, error) {
	sb := strings.Builder{}
	didCreatePr, err := NewRealGitHub().CreatePullRequest(&sb, "work/org/repo1", PullRequest{
		Title:        "some title",
		Body:         "some body",
		UpstreamRepo: "org/repo1",
		IsDraft:      true,
	})

	return didCreatePr, sb.String(), err
}

func runGetDefaultBranchNameAndCaptureOutput() (string, string, error) {
	sb := strings.Builder{}
	defaultBranchName, err := NewRealGitHub().GetDefaultBranchName(&sb, "work/org1/repo1", "org1/repo1")
	return defaultBranchName, sb.String(), err
}

func runUpdatePrDescriptionAndCaptureOutput() (string, error) {
	sb := strings.Builder{}
	err := NewRealGitHub().UpdatePRDescription(&sb, "work/org/repo1", "new title", "new body")
	return sb.String(), err
}

func runIsForkAndCaptureOutput() (bool, string, error) {
	sb := strings.Builder{}
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	originalGit := g
	g = fakeGit
	defer func() {
		g = originalGit
	}()
	isFork, err := NewRealGitHub().IsFork(&sb, "work/org/repo1")
	return isFork, sb.String(), err
}

func runUserHasOpenUpstreamPRsAndCaptureOutput() (bool, string, error) {
	sb := strings.Builder{}
	hasOpenPRs, err := NewRealGitHub().UserHasOpenUpstreamPRs(&sb, "org/repo1")
	return hasOpenPRs, sb.String(), err
}

func runGetOriginRepoNameAndCaptureOutput() (string, string, error) {
	sb := strings.Builder{}
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit
	_, err := NewRealGitHub().GetOriginRepoName(&sb, "work/org/repo1")
	return "", sb.String(), err
}

func newIsForkFakeExecutor(response string, responseErr error) *executor.FakeExecutor {
	return executor.NewFakeExecutor(
		func(string, string, ...string) error { return nil },
		func(_ string, _ string, args ...string) (string, error) {
			switch {
			case isOriginRepoLookup(args):
				return "org/repo1", nil
			case isIsForkLookup(args):
				if responseErr != nil {
					return "", responseErr
				}
				return response, nil
			default:
				return "", nil
			}
		},
	)
}

func isOriginRepoLookup(args []string) bool {
	return len(args) >= 7 && args[0] == "repo" && args[1] == "view" && args[3] == "--json" && args[4] == "nameWithOwner"
}

func isIsForkLookup(args []string) bool {
	return len(args) >= 5 && args[0] == "repo" && args[1] == "view" && args[3] == "--json" && args[4] == "isFork"
}
