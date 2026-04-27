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
	"fmt"
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
		{"user_can_push", "org/repo1"},
		{"clone", "work/org", "org/repo1"},
		{"user_can_push", "org/repo2"},
		{"clone", "work/org", "org/repo2"},
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
		{"fork_and_clone", "work/org", "org/repo1"},
		{"fork_and_clone", "work/org", "org/repo2"},
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
		{"fork_and_clone", "work/org", "org/repo1"},
		{"fork_and_clone", "work/org", "org/repo2"},
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
		{"fork_and_clone", "work/org1", "org1/repo1"},
		{"get_default_branch", "work/org1/repo1", "org1/repo1"},
		{"fork_and_clone", "work/org2", "org2/repo2"},
		{"get_default_branch", "work/org2/repo2", "org2/repo2"},
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
		{"user_can_push", "org1/repo1"},
		{"clone", "work/org1", "org1/repo1"},
		{"user_can_push", "org2/repo2"},
		{"clone", "work/org2", "org2/repo2"},
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
		{"fork_and_clone", "work/org1", "org1/repo1"},
		{"get_default_branch", "work/org1/repo1", "org1/repo1"},
		{"fork_and_clone", "work/org2", "org2/repo2"},
		{"get_default_branch", "work/org2/repo2", "org2/repo2"},
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
		{"fork_and_clone", "work/org1", "org1/repo1"},
		{"get_default_branch", "work/org1/repo1", "org1/repo1"},
		{"fork_and_clone", "work/org2", "org2/repo2"},
		{"get_default_branch", "work/org2/repo2", "org2/repo2"},
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
		{"user_can_push", "org/repo1"},
		{"clone", "work/org", "org/repo1"},
		{"user_can_push", "org/repo2"},
		{"clone", "work/org", "org/repo2"},
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
		{"user_can_push", "orgA/repo1"},
		{"clone", "work/orgA", "orgA/repo1"},
		{"user_can_push", "orgB/repo2"},
		{"clone", "work/orgB", "orgB/repo2"},
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
		{"user_can_push", "mygitserver.com/orgA/repo1"},
		{"clone", "work/orgA", "mygitserver.com/orgA/repo1"},
		{"user_can_push", "orgB/repo2"},
		{"clone", "work/orgB", "orgB/repo2"},
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
		{"fork_and_clone", "work/org", "org/repo2"},
		{"get_default_branch", "work/org/repo2", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org/repo2", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org/repo2", "upstream", "main"},
	})
}

func TestItForksIfUserHasNoPushPermission(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsPushable:
			return false, nil
		case github.ForkAndClone:
			return true, nil
		case github.GetDefaultBranchName:
			return true, nil
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected call")
	})

	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCloneCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Forking and cloning org/repo1")
	assert.Contains(t, out, "Forking and cloning org/repo2")
	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"user_can_push", "org/repo1"},
		{"fork_and_clone", "work/org", "org/repo1"},
		{"get_default_branch", "work/org/repo1", "org/repo1"},
		{"user_can_push", "org/repo2"},
		{"fork_and_clone", "work/org", "org/repo2"},
		{"get_default_branch", "work/org/repo2", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org/repo1", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org/repo1", "upstream", "main"},
		{"checkout", "work/org/repo2", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org/repo2", "upstream", "main"},
	})
}

func TestItForksIfPermissionsCheckFails(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsPushable:
			return false, errors.New("synthetic error")
		case github.ForkAndClone:
			return true, nil
		case github.GetDefaultBranchName:
			return true, nil
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected call")
	})

	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	out, err := runCloneCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Unable to determine if we can push to org/repo1: synthetic error")
	assert.Contains(t, out, "Forking and cloning org/repo1")
	assert.Contains(t, out, "Unable to determine if we can push to org/repo2: synthetic error")
	assert.Contains(t, out, "Forking and cloning org/repo2")
	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"user_can_push", "org/repo1"},
		{"fork_and_clone", "work/org", "org/repo1"},
		{"get_default_branch", "work/org/repo1", "org/repo1"},
		{"user_can_push", "org/repo2"},
		{"fork_and_clone", "work/org", "org/repo2"},
		{"get_default_branch", "work/org/repo2", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"checkout", "work/org/repo1", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org/repo1", "upstream", "main"},
		{"checkout", "work/org/repo2", testsupport.Pwd()},
		{"pull", "--ff-only", "work/org/repo2", "upstream", "main"},
	})
}

func TestFromPRsAndReposAreMutuallyExclusive(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaign(false)
	assert.NoError(t, os.WriteFile("prs.txt", []byte("org/repo#1\n"), 0o644))
	assert.NoError(t, os.WriteFile("my-custom-repos.txt", []byte(""), 0o644))

	// Pass flags via SetArgs so cobra's Changed() detection works — setting
	// the package-level vars directly doesn't mark the flag as user-provided.
	out, err := runCloneCommandArgs([]string{"--from-prs", "prs.txt", "--repos", "my-custom-repos.txt"})
	assert.NoError(t, err)
	assert.Contains(t, out, "mutually exclusive")
}

func TestCloneFromPRsHappyPath(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsPushable, github.Clone, github.CheckoutPR:
			return true, nil
		default:
			return false, errors.New("unexpected command " + fmt.Sprint(command))
		}
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("unexpected")
	})
	gh = fakeGitHub

	fakeGit := git.NewAlwaysSucceedsFakeGit()
	// After CheckoutPR, the fake returns the PR's head branch per repo.
	fakeGit.SetCurrentBranchName("work/org/repo1", "feat/fix-1")
	fakeGit.SetCurrentBranchName("work/org/repo2", "feat/fix-2")
	g = fakeGit

	testsupport.PrepareTempCampaign(false) // empty repos.txt (just a few template comments)
	assert.NoError(t, os.WriteFile("prs.txt", []byte("org/repo1#1\norg/repo2#2\n"), 0o644))

	out, err := runCloneCommandArgs([]string{"--from-prs", "prs.txt"})
	assert.NoError(t, err)
	assert.Contains(t, out, "Assimilating PR #1 for org/repo1")
	assert.Contains(t, out, "Assimilating PR #2 for org/repo2")
	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	reposContent, err := os.ReadFile("repos.txt")
	assert.NoError(t, err)
	assert.Contains(t, string(reposContent), "org/repo1 # branch=feat/fix-1")
	assert.Contains(t, string(reposContent), "org/repo2 # branch=feat/fix-2")

	calls := fakeGitHub.Calls()
	assertContainsCall(t, calls, []string{"checkout_pr", "work/org/repo1", "1"})
	assertContainsCall(t, calls, []string{"checkout_pr", "work/org/repo2", "2"})
}

func TestCloneFromPRsKeysBranchesByHostOrgRepoForGHE(t *testing.T) {
	// For a GHE PR URL, repos.txt stores `host/org/repo` and the branch
	// annotation must be keyed on that same full identifier — otherwise the
	// upsert won't find the existing line and we'd either append a duplicate
	// or lose the host prefix.
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsPushable, github.Clone, github.CheckoutPR:
			return true, nil
		default:
			return false, errors.New("unexpected command")
		}
	}, func(workingDir string) (interface{}, error) { return nil, nil })
	gh = fakeGitHub

	fakeGit := git.NewAlwaysSucceedsFakeGit()
	fakeGit.SetCurrentBranchName("work/org/repo1", "feat/fix")
	g = fakeGit

	testsupport.PrepareTempCampaign(false) // empty repos.txt
	assert.NoError(t, os.WriteFile("prs.txt", []byte("https://my-ghe.example/org/repo1/pull/1\n"), 0o644))

	_, err := runCloneCommandArgs([]string{"--from-prs", "prs.txt"})
	assert.NoError(t, err)

	// repos.txt must carry the host-qualified repo name, not `org/repo1`.
	reposContent, err := os.ReadFile("repos.txt")
	assert.NoError(t, err)
	assert.Contains(t, string(reposContent), "my-ghe.example/org/repo1 # branch=feat/fix")
}

func TestCloneFromPRsFailsOnConflictingExistingAnnotation(t *testing.T) {
	fakeGitHub := github.NewFakeGitHub(func(command github.Command, args []string) (bool, error) {
		switch command {
		case github.IsPushable, github.Clone, github.CheckoutPR:
			return true, nil
		default:
			return false, errors.New("unexpected")
		}
	}, func(workingDir string) (interface{}, error) { return nil, nil })
	gh = fakeGitHub

	fakeGit := git.NewAlwaysSucceedsFakeGit()
	// PR checkout lands us on a different branch than what's already
	// annotated in repos.txt — that's the conflict we want to detect.
	fakeGit.SetCurrentBranchName("work/org/repo1", "new-branch")
	g = fakeGit

	testsupport.PrepareTempCampaign(false, "org/repo1 # branch=old-branch")
	assert.NoError(t, os.WriteFile("prs.txt", []byte("org/repo1#1\n"), 0o644))

	originalRepos, _ := os.ReadFile("repos.txt")

	out, err := runCloneCommandArgs([]string{"--from-prs", "prs.txt"})
	assert.NoError(t, err)
	assert.Contains(t, out, "conflicting")

	// UpsertBranchAnnotations is atomic — file must be byte-identical.
	afterRepos, _ := os.ReadFile("repos.txt")
	assert.Equal(t, string(originalRepos), string(afterRepos))
}

func assertContainsCall(t *testing.T, calls [][]string, want []string) {
	t.Helper()
	for _, c := range calls {
		if len(c) != len(want) {
			continue
		}
		match := true
		for i := range c {
			if c[i] != want[i] {
				match = false
				break
			}
		}
		if match {
			return
		}
	}
	t.Errorf("expected call %v not found in %v", want, calls)
}

func runCloneCommand() (string, error) {
	return runCloneCommandArgs(nil)
}

func runCloneCommandWithFork() (string, error) {
	cmd := NewCloneCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	forceFork = true
	prsFile = ""
	repoFile = "repos.txt"
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func runCloneCommandArgs(args []string) (string, error) {
	cmd := NewCloneCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	forceFork = false
	prsFile = ""
	repoFile = "repos.txt"
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
