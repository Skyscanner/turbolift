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

package foreach

import (
	"bytes"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/skyscanner/turbolift/internal/testsupport"
)

func TestItRejectsEmptyArgs(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand([]string{}...)
	assert.Error(t, err, "Expected an error to be returned")
	assert.Contains(t, out, "Usage")

	fakeExecutor.AssertCalledWith(t, [][]string{})
}

func TestItRejectsCommandWithoutDashes(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("some", "command")
	assert.Error(t, err, "Expected an error to be returned")
	assert.Contains(t, out, "Usage")

	fakeExecutor.AssertCalledWith(t, [][]string{})
}

func TestItRunsCommandWithoutShellAgainstWorkingCopies(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "some", "command"},
		{"work/org/repo2", "some", "command"},
	})
}

func TestItRunsCommandWithSpacesAgainstWorkingCopied(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("--", "some", "command", "with spaces")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "2 OK, 0 skipped")
	assert.Contains(t, out,
		"Executing { some command 'with spaces' } in work/org/repo1",
		"It should format the executed command accurately")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "some", "command", "with spaces"},
		{"work/org/repo2", "some", "command", "with spaces"},
	})
}

func TestItSkipsMissingWorkingCopies(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	_ = os.Remove("work/org/repo2")

	out, err := runCommand("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "1 OK, 1 skipped")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "some", "command"},
	})
}

func TestItContinuesOnAndRecordsFailures(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed with errors")
	assert.Contains(t, out, "0 OK, 0 skipped, 2 errored")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "some", "command"},
		{"work/org/repo2", "some", "command"},
	})
}

func TestHelpFlagReturnsUsage(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("--help", "--", "command1")
	t.Log(out)
	assert.NoError(t, err)
	// should return usage
	assert.Contains(t, out, "Usage:")
	assert.Contains(t, out, "foreach [flags] -- COMMAND [ARGUMENT...]")
	assert.Contains(t, out, "Flags:")
	assert.Contains(t, out, "help for foreach")

	// nothing should have been called
	fakeExecutor.AssertCalledWith(t, [][]string{})
}

func TestFormatArguments(t *testing.T) {
	// Don't go too heavy here. We are not seeking to exhaustively test
	// shellescape. We just want to make sure formatArguments works.
	var tests = []struct {
		input    []string
		expected string
		title    string
	}{
		{[]string{""}, `''`, "Empty arg should be quoted"},
		{[]string{"one two"}, `'one two'`, "Arg with space should be quoted"},
		{[]string{"one"}, `one`, "Plain arg should not need quotes"},
		{[]string{}, ``, "Empty arg list should give empty string"},
		{[]string{"x", "", "y y"}, `x '' 'y y'`, "Args should be separated with spaces"},
	}
	for _, test := range tests {
		actual := formatArguments(test.input)
		assert.Equal(t, actual, test.expected, test.title)
	}
}

func TestItCreatesLogFiles(t *testing.T) {
	fakeExecutor := executor.NewAlternatingSuccessFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "1 OK, 0 skipped, 1 errored")

	// Logs should describe where output was written
	r := regexp.MustCompile(`Logs for all executions have been stored under (.+)`)
	matches := r.FindStringSubmatch(out)
	assert.Len(t, matches, 2, "Expected to find the log directory path")
	path := matches[1]

	// check that expected static directories and files exist
	_, err = os.Stat(path)
	assert.NoError(t, err, "Expected the log directory to exist")

	_, err = os.Stat(path + "/successful")
	assert.NoError(t, err, "Expected the successful log directory to exist")

	_, err = os.Stat(path + "/failed")
	assert.NoError(t, err, "Expected the failure log directory to exist")

	_, err = os.Stat(path + "/successful/repos.txt")
	assert.NoError(t, err, "Expected the successful repos.txt file to exist")

	_, err = os.Stat(path + "/failed/repos.txt")
	assert.NoError(t, err, "Expected the failure repos.txt file to exist")

	// check that the expected logs files exist
	_, err = os.Stat(path + "/successful/org/repo1/logs.txt")
	assert.NoError(t, err, "Expected the successful log file for org/repo1 to exist")

	_, err = os.Stat(path + "/failed/org/repo2/logs.txt")
	assert.NoError(t, err, "Expected the failure log file for org/repo2 to exist")
}

func TestItRunsAgainstSuccessfulReposOnly(t *testing.T) {
	fakeExecutor := executor.NewAlternatingSuccessFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repo3")
	err := setUpSymlink()
	if err != nil {
		t.Errorf("Error setting up symlink: %s", err)
	}
	defer os.RemoveAll("mock_output")

	out, err := runCommandReposSuccessful("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "1 OK, 0 skipped, 1 errored")
	assert.Contains(t, out, "org/repo1")
	assert.Contains(t, out, "org/repo3")
	assert.NotContains(t, out, "org/repo2")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "some", "command"},
		{"work/org/repo3", "some", "command"},
	})
}

func TestItRunsAgainstFailedReposOnly(t *testing.T) {
	fakeExecutor := executor.NewAlternatingSuccessFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repo3")
	err := setUpSymlink()
	if err != nil {
		t.Errorf("Error setting up symlink: %s", err)
	}
	defer os.RemoveAll("mock_output")

	out, err := runCommandReposFailed("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "1 OK, 0 skipped, 1 errored")
	assert.Contains(t, out, "org/repo1")
	assert.Contains(t, out, "org/repo3")
	assert.NotContains(t, out, "org/repo2")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "some", "command"},
		{"work/org/repo3", "some", "command"},
	})
}

func TestItCreatesSymlinksSuccessfully(t *testing.T) {
	fakeExecutor := executor.NewAlternatingSuccessFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repo3")

	out, err := runCommand("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "2 OK, 0 skipped, 1 errored")

	resultsDir, err := os.Readlink(".turbolift_previous_results")
	if err != nil {

		t.Errorf("Error reading symlink: %s", err)
	}

	successfulRepoFile := path.Join(resultsDir, "successful", "repos.txt")
	successfulRepos, err := os.ReadFile(successfulRepoFile)
	if err != nil {
		t.Errorf("Error reading successful repos: %s", err)
	}
	assert.Contains(t, string(successfulRepos), "org/repo1")
	assert.Contains(t, string(successfulRepos), "org/repo3")
	assert.NotContains(t, string(successfulRepos), "org/repo2")

	failedRepoFile := path.Join(resultsDir, "failed", "repos.txt")
	failedRepos, err := os.ReadFile(failedRepoFile)
	if err != nil {
		t.Errorf("Error reading failed repos: %s", err)
	}
	assert.Contains(t, string(failedRepos), "org/repo2")
	assert.NotContains(t, string(failedRepos), "org/repo1")
	assert.NotContains(t, string(failedRepos), "org/repo3")
}

func TestItRunsAgainstCustomReposFile(t *testing.T) {
	fakeExecutor := executor.NewAlternatingSuccessFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repo3")
	testsupport.CreateAnotherRepoFile("custom_repofile.txt", "org/repo1", "org/repo3")

	out, err := runCommandReposCustom("--", "some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "1 OK, 0 skipped, 1 errored")
	assert.Contains(t, out, "org/repo1")
	assert.Contains(t, out, "org/repo3")
	assert.NotContains(t, out, "org/repo2")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "some", "command"},
		{"work/org/repo3", "some", "command"},
	})
}

func TestItDoesNotAllowMultipleReposArguments(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repo3")

	_, err := runCommandReposMultiple("--", "some", "command")
	assert.Error(t, err, "only one repositories flag or option may be specified: either --successful; --failed; or --repos <file>")

	fakeExecutor.AssertCalledWith(t, [][]string{})
}

func setUpSymlink() error {
	err := os.MkdirAll("mock_output/successful", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll("mock_output/failed", 0755)
	if err != nil {
		return err
	}
	err = os.Symlink("mock_output", ".turbolift_previous_results")
	if err != nil {
		return err
	}
	_, err = os.Create("mock_output/successful/repos.txt")
	if err != nil {
		return err
	}
	_, err = os.Create("mock_output/failed/repos.txt")
	if err != nil {
		return err
	}
	repos := []string{"org/repo1", "org/repo3"}
	delimitedList := strings.Join(repos, "\n")
	_ = os.WriteFile("mock_output/successful/repos.txt", []byte(delimitedList), os.ModePerm|0o644)
	_ = os.WriteFile("mock_output/failed/repos.txt", []byte(delimitedList), os.ModePerm|0o644)
	return nil
}

func runCommand(args ...string) (string, error) {
	cmd := NewForeachCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runCommandReposSuccessful(args ...string) (string, error) {
	cmd := NewForeachCmd()
	successful = true
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runCommandReposFailed(args ...string) (string, error) {
	cmd := NewForeachCmd()
	failed = true
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runCommandReposCustom(args ...string) (string, error) {
	cmd := NewForeachCmd()
	repoFile = "custom_repofile.txt"
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runCommandReposMultiple(args ...string) (string, error) {
	cmd := NewForeachCmd()
	successful = true
	repoFile = "custom_repofile.txt"
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
