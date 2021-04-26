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

package clone

import (
	"bytes"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestItAbortsIfReposFileNotFound(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false)
	err := os.Remove("repos.txt")
	if err != nil {
		panic(err)
	}

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "❌ Reading campaign data")

	fakeGitHub.AssertCalledWith(t, [][]string{})
	fakeGit.AssertCalledWith(t, [][]string{})
}

func TestItLogsCloneErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "❌ Forking and cloning org/repo1")
	assert.Contains(t, out, "❌ Forking and cloning org/repo2")
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

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "❌ Creating branch")
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

func TestItClonesReposFoundInReposFile(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)

	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo1"},
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

	_, err := runCommand()
	assert.NoError(t, err)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/orgA", "orgA/repo1"},
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

	_, err := runCommand()
	assert.NoError(t, err)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/orgA", "mygitserver.com/orgA/repo1"},
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
	_ = os.MkdirAll(path.Join("work", "org", "repo1"), os.ModeDir|0755)

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "⚠️  Forking and cloning org/repo1")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org/repo2", testsupport.Pwd()},
	})
}

func runCommand() (string, error) {
	cmd := NewCloneCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()

	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
