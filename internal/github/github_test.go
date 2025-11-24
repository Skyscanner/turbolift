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
	calls := 0
	fakeExecutor := executor.NewFakeExecutor(func(workingDir string, name string, args ...string) error {
		return nil
	}, func(workingDir string, name string, args ...string) (string, error) {
		calls++
		if calls == 1 {
			return "", nil // label create succeeds
		}
		return "", errors.New("synthetic error") // pr create fails
	})
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrAndCaptureOutput()
	assert.Error(t, err)
	assert.False(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "label", "create", "turbolift", "--repo", "org/repo1", "--color", turboliftLabelColor, "--description", turboliftLabelDescription},
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1", "--label", "turbolift"},
	})
}

func TestItReturnsFalseAndNilErrorOnNoOpCreatePr(t *testing.T) {
	calls := 0
	fakeExecutor := executor.NewFakeExecutor(func(workingDir string, name string, args ...string) error {
		return nil
	}, func(workingDir string, name string, args ...string) (string, error) {
		calls++
		if calls == 1 {
			return "", nil // label create succeeds
		}
		return "... GraphQL error: No commits between A and B ...", errors.New("synthetic error")
	})
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrAndCaptureOutput()
	assert.NoError(t, err)
	assert.False(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "label", "create", "turbolift", "--repo", "org/repo1", "--color", turboliftLabelColor, "--description", turboliftLabelDescription},
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1", "--label", "turbolift"},
	})
}

func TestItSuccessfulCreatesADraftPr(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreateDraftPrAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "label", "create", "turbolift", "--repo", "org/repo1", "--color", turboliftLabelColor, "--description", turboliftLabelDescription},
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1", "--label", "turbolift", "--draft"},
	})
}

func TestItReturnsTrueAndNilErrorOnSuccessfulCreatePr(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "label", "create", "turbolift", "--repo", "org/repo1", "--color", turboliftLabelColor, "--description", turboliftLabelDescription},
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1", "--label", "turbolift"},
	})
}

func TestItCreatesPrWithoutLabelsWhenNoneProvided(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrWithoutLabelsAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1"},
	})
}

func TestItIgnoresExistingLabelError(t *testing.T) {
	calls := 0
	fakeExecutor := executor.NewFakeExecutor(func(workingDir string, name string, args ...string) error {
		return nil
	}, func(workingDir string, name string, args ...string) (string, error) {
		calls++
		if calls == 1 {
			return "label already exists", errors.New("synthetic error")
		}
		return "", nil
	})
	execInstance = fakeExecutor

	didCreatePr, _, err := runCreatePrAndCaptureOutput()
	assert.NoError(t, err)
	assert.True(t, didCreatePr)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "gh", "label", "create", "turbolift", "--repo", "org/repo1", "--color", turboliftLabelColor, "--description", turboliftLabelDescription},
		{"work/org/repo1", "gh", "pr", "create", "--title", "some title", "--body", "some body", "--repo", "org/repo1", "--label", "turbolift"},
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
		Labels:       []string{TurboliftLabel},
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
		Labels:       []string{TurboliftLabel},
	})

	return didCreatePr, sb.String(), err
}

func runCreatePrWithoutLabelsAndCaptureOutput() (bool, string, error) {
	sb := strings.Builder{}
	didCreatePr, err := NewRealGitHub().CreatePullRequest(&sb, "work/org/repo1", PullRequest{
		Title:        "some title",
		Body:         "some body",
		UpstreamRepo: "org/repo1",
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
