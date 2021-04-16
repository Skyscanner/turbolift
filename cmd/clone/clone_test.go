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

	testsupport.PrepareTempCampaignDirectory()
	err := os.Remove("repos.txt")
	if err != nil {
		panic(err)
	}

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Error when reading campaign directory")

	fakeGitHub.AssertCalledWith(t, [][]string{})
	fakeGit.AssertCalledWith(t, [][]string{})
}

func TestItLogsCloneErrorsButContinuesToTryAll(t *testing.T) {
	fakeGitHub := github.NewAlwaysFailsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysFailsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Error when cloning org/repo1")
	assert.Contains(t, out, "Error when cloning org/repo2")
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

	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Error when creating branch")
	assert.Contains(t, out, "turbolift clone completed with errors")
	assert.Contains(t, out, "2 repos errored")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo1"},
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"work/org/repo1", testsupport.Pwd()},
		{"work/org/repo2", testsupport.Pwd()},
	})

}

func TestItClonesReposFoundInReposFile(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)

	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo1"},
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"work/org/repo1", testsupport.Pwd()},
		{"work/org/repo2", testsupport.Pwd()},
	})
}

func TestItClonesReposInMultipleOrgs(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory("orgA/repo1", "orgB/repo2")

	_, err := runCommand()
	assert.NoError(t, err)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/orgA", "orgA/repo1"},
		{"work/orgB", "orgB/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"work/orgA/repo1", testsupport.Pwd()},
		{"work/orgB/repo2", testsupport.Pwd()},
	})
}

func TestItClonesReposFromOtherHosts(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory("mygitserver.com/orgA/repo1", "orgB/repo2")

	_, err := runCommand()
	assert.NoError(t, err)

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/orgA", "mygitserver.com/orgA/repo1"},
		{"work/orgB", "orgB/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"work/orgA/repo1", testsupport.Pwd()},
		{"work/orgB/repo2", testsupport.Pwd()},
	})
}

func TestItSkipsCloningIfAWorkingCopyAlreadyExists(t *testing.T) {
	fakeGitHub := github.NewAlwaysSucceedsFakeGitHub()
	gh = fakeGitHub
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")
	_ = os.MkdirAll(path.Join("work", "org", "repo1"), os.ModeDir|0755)

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Not cloning org/repo1")

	fakeGitHub.AssertCalledWith(t, [][]string{
		{"work/org", "org/repo2"},
	})
	fakeGit.AssertCalledWith(t, [][]string{
		{"work/org/repo2", testsupport.Pwd()},
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
