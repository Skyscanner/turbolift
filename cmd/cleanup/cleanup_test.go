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

package cleanup

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/testsupport"
)

func TestItWritesDeletableForksToFile(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsFork:
			return true, nil
		case github.UserHasOpenUpstreamPRs:
			if args[1] == "org/repo1" {
				return false, nil
			} else {
				return true, nil
			}
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return "org/repo", nil
	})

	gh = fakeGitHub

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCleanupCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift cleanup completed (2 forks checked, 0 non-forks skipped)")
	assert.Contains(t, out, "If you wish to delete these forks, run the following command:")
	assert.FileExists(t, cleanupFile)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"is_fork", "work/org/repo1"},
		{"user_has_open_upstream_prs", "org/repo1"},
		{"get_origin_repo_name", "work/org/repo1"},
		{"is_fork", "work/org/repo2"},
		{"user_has_open_upstream_prs", "org/repo2"},
	})
}

func TestItWritesNothingWhenNoForksAreDeletable(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsFork:
			return true, nil
		case github.UserHasOpenUpstreamPRs:
			return true, nil
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected call")
	})

	gh = fakeGitHub

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCleanupCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift cleanup completed (2 forks checked, 0 non-forks skipped)")
	assert.Contains(t, out, "All forks used in this campaign appear to have an open upstream PR.")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"is_fork", "work/org/repo1"},
		{"user_has_open_upstream_prs", "org/repo1"},
		{"is_fork", "work/org/repo2"},
		{"user_has_open_upstream_prs", "org/repo2"},
	})
}

func TestItSkipsNonForksButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsFork:
			return false, nil
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected call")
	})

	gh = fakeGitHub

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCleanupCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift cleanup completed (0 forks checked, 2 non-forks skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"is_fork", "work/org/repo1"},
		{"is_fork", "work/org/repo2"},
	})
}

func TestItWarnsOnErrorButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsFork:
			return true, nil
		case github.UserHasOpenUpstreamPRs:
			return false, errors.New("synthetic error")
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected call")
	})

	gh = fakeGitHub

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCleanupCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "turbolift cleanup completed with errors (0 forks checked, 0 non-forks skipped, 2 errored)")
	assert.Contains(t, out, "Please check errors above and fix if necessary")
	assert.FileExists(t, cleanupFile)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"is_fork", "work/org/repo1"},
		{"user_has_open_upstream_prs", "org/repo1"},
		{"is_fork", "work/org/repo2"},
		{"user_has_open_upstream_prs", "org/repo2"},
	})
}

func runCleanupCommand() (string, error) {
	cmd := NewCleanupCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
