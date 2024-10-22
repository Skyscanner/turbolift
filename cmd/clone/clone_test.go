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

package clone

import (
	"bytes"
	"errors"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/testsupport"
)

func TestItAbortsIfReposFileNotFound(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	_ = testsupport.PrepareTempCampaign(false)
	err := os.Remove("repos.txt")
	if err != nil {
		panic(err)
	}

	out, err := runCloneCommandWithFork()
	assert.NoError(t, err)
	assert.Contains(t, out, "Reading campaign data")

	fakeGitHub.AssertCalledWith(t, [][]string{})
	fakeGit.AssertCalledWith(t, [][]string{})
}

func TestItLogsCloneErrorsButContinuesToTryAll(t *testing.T) {
	// this fakeGithub will tell the caller that the repo is pushable, but will fail to clone it
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsPushable:
			return true, nil
		case github.Clone:
			return false, errors.New("synthetic error")
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected call")
	})

	// fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCloneCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Cloning org/repo1")
	assert.Contains(t, out, "Cloning org/repo2")
	assert.Contains(t, out, "turbolift clone completed with errors")
	assert.Contains(t, out, "2 repos errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1"},
		{"work/org", "org/repo1"},
		{"work/org/repo2"},
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{})
}

func TestItLogsForkAndCloneErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCloneCommandWithFork()
	assert.NoError(t, err)
	assert.Contains(t, out, "Forking and cloning org/repo1")
	assert.Contains(t, out, "Forking and cloning org/repo2")
	assert.Contains(t, out, "turbolift clone completed with errors")
	assert.Contains(t, out, "2 repos errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo1"},
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{})
}

func TestItLogsCheckoutErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCloneCommandWithFork()
	assert.NoError(t, err)
	assert.Contains(t, out, "Creating branch")
	assert.Contains(t, out, "turbolift clone completed with errors")
	assert.Contains(t, out, "2 repos errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo1"},
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org/repo1", testsupport.Pwd()},
		{"checkout", "work/org/repo2", testsupport.Pwd()},
	})
}

func TestItPullsFromUpstreamWhenCloningWithFork(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org1/repo1", "org2/repo2")

	out, err := runCloneCommandWithFork()
	assert.NoError(t, err)
	assert.Contains(t, out, "Pulling latest changes from org1/repo1")
	assert.Contains(t, out, "Pulling latest changes from org2/repo2")
	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org1", "org1/repo1"},
		{"work/org1/repo1", "org1/repo1"},
		{"work/org2", "org2/repo2"},
		{"work/org2/repo2", "org2/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org1/repo1", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org1/repo1", "upstream", "main"},
		{"checkout", "work/org2/repo2", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org2/repo2", "upstream", "main"},
	})
}

func TestItDoesNotPullFromUpstreamWhenCloningWithoutFork(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsPushable, github.Clone:
			return true, nil
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected call")
	})

	// fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org1/repo1", "org2/repo2")

	out, err := runCloneCommand()
	assert.NoError(t, err)
	assert.NotContains(t, out, "Pulling latest changes from org1/repo1")
	assert.NotContains(t, out, "Pulling latest changes from org2/repo2")
	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org1/repo1"},
		{"work/org1", "org1/repo1"},
		{"work/org2/repo2"},
		{"work/org2", "org2/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org1/repo1", testsupport.Pwd()},
		{"checkout", "work/org2/repo2", testsupport.Pwd()},
	})
}

func TestItLogsDefaultBranchErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsOnGetDefaultBranchFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org1/repo1", "org2/repo2")
	out, err := runCloneCommandWithFork()
	assert.NoError(t, err)
	assert.Contains(t, out, "Pulling latest changes from org1/repo1")
	assert.Contains(t, out, "Pulling latest changes from org2/repo2")
	assert.Contains(t, out, "turbolift clone completed with errors")
	assert.Contains(t, out, "2 repos errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org1", "org1/repo1"},
		{"work/org1/repo1", "org1/repo1"},
		{"work/org2", "org2/repo2"},
		{"work/org2/repo2", "org2/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org1/repo1", testsupport.Pwd()},
		{"checkout", "work/org2/repo2", testsupport.Pwd()},
	})
}

func TestItLogsPullErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsOnPullFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org1/repo1", "org2/repo2")
	out, err := runCloneCommandWithFork()
	assert.NoError(t, err)
	assert.Contains(t, out, "Pulling latest changes from org1/repo1")
	assert.Contains(t, out, "Pulling latest changes from org2/repo2")
	assert.Contains(t, out, "We weren't able to pull the latest upstream changes into your fork of org1/repo1")
	assert.Contains(t, out, "We weren't able to pull the latest upstream changes into your fork of org2/repo2")
	assert.Contains(t, out, "turbolift clone completed with errors")
	assert.Contains(t, out, "2 repos errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org1", "org1/repo1"},
		{"work/org1/repo1", "org1/repo1"},
		{"work/org2", "org2/repo2"},
		{"work/org2/repo2", "org2/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org1/repo1", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org1/repo1", "upstream", "main"},
		{"checkout", "work/org2/repo2", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org2/repo2", "upstream", "main"},
	})
}

func TestItClonesReposFoundInReposFile(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCloneCommand()
	assert.NoError(t, err)

	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org/repo1"},
		{"work/org", "org/repo1"},
		{"work/org/repo2"},
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org/repo1", testsupport.Pwd()},
		{"checkout", "work/org/repo2", testsupport.Pwd()},
	})
}

func TestItClonesReposInMultipleOrgs(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "orgA/repo1", "orgB/repo2")

	_, err := runCloneCommand()
	assert.NoError(t, err)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/orgA/repo1"},
		{"work/orgA", "orgA/repo1"},
		{"work/orgB/repo2"},
		{"work/orgB", "orgB/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/orgA/repo1", testsupport.Pwd()},
		{"checkout", "work/orgB/repo2", testsupport.Pwd()},
	})
}

func TestItClonesReposFromOtherHosts(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "mygitserver.com/orgA/repo1", "orgB/repo2")

	_, err := runCloneCommand()
	assert.NoError(t, err)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/orgA/repo1"},
		{"work/orgA", "mygitserver.com/orgA/repo1"},
		{"work/orgB/repo2"},
		{"work/orgB", "orgB/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/orgA/repo1", testsupport.Pwd()},
		{"checkout", "work/orgB/repo2", testsupport.Pwd()},
	})
}

func TestItSkipsCloningIfAWorkingCopyAlreadyExists(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")
	_ = os.MkdirAll(path.Join("work", "org", "repo1"), os.ModeDir|0o755)

	out, err := runCloneCommandWithFork()
	assert.NoError(t, err)
	assert.Contains(t, out, "Forking and cloning org/repo1")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo2"},
		{"work/org/repo2", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org/repo2", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org/repo2", "upstream", "main"},
	})
}

func runCloneCommand() (string, error) {
	cmd := NewCloneCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	forceFork = false
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runCloneCommandWithFork() (string, error) {
	cmd := NewCloneCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	forceFork = true
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
