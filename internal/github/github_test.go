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

// TestGetPR_ForkPRMatchesByExactBranch exercises the fork-PR lookup path
// where `gh pr status` returns multiple PRs in createdBy. Under a naive
// `strings.HasSuffix` match, a longer branch name ending with the one we're
// looking for would match first (e.g. "foo-feat/x" contains suffix "feat/x").
// The fix is to require either an exact match or a `user:branch` fork-style
// suffix — i.e. the branch preceded by ':'.
func TestGetPR_ForkPRMatchesByExactBranch(t *testing.T) {
	// createdBy[0] has HeadRefName "someuser:foo-feat/x" — plain HasSuffix
	// would wrongly match this for branchName "feat/x". createdBy[1] has
	// "otheruser:feat/x" which is the correct fork-style match.
	prStatusJSON := `{
		"currentBranch": null,
		"createdBy": [
			{"number": 1, "headRefName": "someuser:foo-feat/x", "state": "OPEN"},
			{"number": 2, "headRefName": "otheruser:feat/x", "state": "OPEN"}
		],
		"needsReview": []
	}`
	fakeExecutor := executor.NewFakeExecutor(
		func(workingDir string, name string, args ...string) error { return nil },
		func(workingDir string, name string, args ...string) (string, error) {
			return prStatusJSON, nil
		},
	)
	execInstance = fakeExecutor

	sb := strings.Builder{}
	pr, err := NewRealGitHub().GetPR(&sb, "work/org/repo1", "feat/x")
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 2, pr.Number, "should match the PR whose branch is exactly feat/x (preceded by ':'), not the one that merely ends in feat/x")
}
