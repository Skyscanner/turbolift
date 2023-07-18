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

package foreach

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/skyscanner/turbolift/internal/testsupport"
)

func TestParseForEachArgs(t *testing.T) {
	testCases := []struct {
		Name                 string
		Args                 []string
		ExpectedCommand      []string
		ExpectedRepoFileName string
		ExpectedHelpFlag     bool
	}{
		{
			Name:                 "simple command",
			Args:                 []string{"ls", "-l"},
			ExpectedCommand:      []string{"ls", "-l"},
			ExpectedRepoFileName: "repos.txt",
			ExpectedHelpFlag:     false,
		},
		{
			Name:                 "advanced command",
			Args:                 []string{"sed", "-e", "'s/foo/bar/'", "-e", "'s/bar/baz/'"},
			ExpectedCommand:      []string{"sed", "-e", "'s/foo/bar/'", "-e", "'s/bar/baz/'"},
			ExpectedRepoFileName: "repos.txt",
			ExpectedHelpFlag:     false,
		},
		{
			Name:                 "simple command with repo flag",
			Args:                 []string{"--repos", "test.txt", "ls", "-l"},
			ExpectedCommand:      []string{"ls", "-l"},
			ExpectedRepoFileName: "test.txt",
			ExpectedHelpFlag:     false,
		},
		{
			Name:                 "advanced command with repos flag",
			Args:                 []string{"--repos", "test2.txt", "sed", "-e", "'s/foo/bar/'", "-e", "'s/bar/baz/'"},
			ExpectedCommand:      []string{"sed", "-e", "'s/foo/bar/'", "-e", "'s/bar/baz/'"},
			ExpectedRepoFileName: "test2.txt",
			ExpectedHelpFlag:     false,
		},
		{
			Name:                 "repos flag should only be caught when at the beginning",
			Args:                 []string{"ls", "-l", "--repos", "random.txt"},
			ExpectedCommand:      []string{"ls", "-l", "--repos", "random.txt"},
			ExpectedRepoFileName: "repos.txt",
			ExpectedHelpFlag:     false,
		},
		{
			Name:                 "random flag is not caught",
			Args:                 []string{"--random", "arg", "ls", "-l"},
			ExpectedCommand:      []string{"--random", "arg", "ls", "-l"},
			ExpectedRepoFileName: "repos.txt",
			ExpectedHelpFlag:     false,
		},
		{
			Name:                 "Help flag is triggered",
			Args:                 []string{"--help"},
			ExpectedCommand:      []string{},
			ExpectedRepoFileName: "repos.txt",
			ExpectedHelpFlag:     true,
		},
		{
			Name:                 "Help flag is triggered after the repo one",
			Args:                 []string{"--repos", "example.txt", "--help", "thecommand"},
			ExpectedCommand:      []string{"thecommand"},
			ExpectedRepoFileName: "example.txt",
			ExpectedHelpFlag:     true,
		},
		{
			Name:                 "Help flag is triggered before the repo one",
			Args:                 []string{"--help", "--repos", "example.txt", "newcommand", "anotherarg"},
			ExpectedCommand:      []string{"newcommand", "anotherarg"},
			ExpectedRepoFileName: "example.txt",
			ExpectedHelpFlag:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := parseForeachArgs(tc.Args)
			t.Log(actual)
			assert.EqualValues(t, tc.ExpectedCommand, actual)
			assert.Equal(t, repoFile, tc.ExpectedRepoFileName)
			assert.Equal(t, helpFlag, tc.ExpectedHelpFlag)

			// Cleanup to default repo file name
			repoFile = "repos.txt"
			helpFlag = false
		})
	}
}

func TestItRejectsEmptyArgs(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand([]string{}...)
	assert.Errorf(t, err, "requires at least 1 arg(s), only received 0")
	assert.Contains(t, out, "Usage")

	fakeExecutor.AssertCalledWith(t, [][]string{})
}

func TestItRunsCommandInShellAgainstWorkingCopies(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "2 OK, 0 skipped")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", userShell(), "-c", "some command"},
		{"work/org/repo2", userShell(), "-c", "some command"},
	})
}

func TestItSkipsMissingWorkingCopies(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	_ = os.Remove("work/org/repo2")

	out, err := runCommand("some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed")
	assert.Contains(t, out, "1 OK, 1 skipped")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", userShell(), "-c", "some command"},
	})
}

func TestItContinuesOnAndRecordsFailures(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("some", "command")
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift foreach completed with errors")
	assert.Contains(t, out, "0 OK, 0 skipped, 2 errored")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", userShell(), "-c", "some command"},
		{"work/org/repo2", userShell(), "-c", "some command"},
	})
}

func TestHelpFlagReturnsUsage(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")

	out, err := runCommand("--help", "command1")
	t.Log(out)
	assert.NoError(t, err)
	// should return usage
	assert.Contains(t, out, "Usage:")
	assert.Contains(t, out, "foreach SHELL_COMMAND [flags]")
	assert.Contains(t, out, "Flags:")
	assert.Contains(t, out, "help for foreach")

	// nothing should have been called
	fakeExecutor.AssertCalledWith(t, [][]string{})
}

func userShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "sh"
	}
	return shell
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
