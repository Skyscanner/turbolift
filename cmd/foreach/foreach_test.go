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
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

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
