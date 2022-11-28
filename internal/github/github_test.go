/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/skyscanner/turbolift/cmd/flags"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/stretchr/testify/assert"
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

func TestItReturnsNilErrorOnSuccessfulForkWithDryRun(t *testing.T) {
	flags.DryRun = true
	t.Cleanup(func() {
		flags.DryRun = false
	})

	_ = executor.NewRealExecutor()

	output, err := runForkAndCloneAndCaptureOutput()
	assert.NoError(t, err)
	assert.Equal(t, "Dry-run mode: gh [repo fork --clone=true org/repo1]. Working dir: work/org", output)
}

func TestItReturnsNilErrorOnSuccessfulPRCreationWithDryRun(t *testing.T) {
	flags.DryRun = true
	t.Cleanup(func() {
		flags.DryRun = false
	})

	_ = executor.NewRealExecutor()
	didCreatePr, output, err := runCreatePrAndCaptureOutput()
	assert.NoError(t, err)
	assert.False(t, didCreatePr)
	assert.Equal(t, "Dry-run mode: gh [pr create --title some title --body some body --repo org/repo1]. Working dir: work/org/repo1", output)
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
