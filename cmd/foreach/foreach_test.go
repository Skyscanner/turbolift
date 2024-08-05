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
	"regexp"
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
