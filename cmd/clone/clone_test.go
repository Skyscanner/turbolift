package clone

import (
	"bytes"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestItAbortsIfReposFileNotFound(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaignDirectory()
	err := os.Remove("repos.txt")
	if err != nil {
		panic(err)
	}

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Error when reading campaign directory")

	fakeExecutor.AssertCalledWith(t, [][]string{})
}

func TestItLogsCloneErrorsButContinuesToTryAll(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Error when cloning org/repo1")
	assert.Contains(t, out, "Error when cloning org/repo2")
	assert.Contains(t, out, "turbolift clone completed with errors")
	assert.Contains(t, out, "2 repos errored")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo2"},
	})
}

func TestItClonesReposFoundInReposFile(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")

	out, err := runCommand()
	assert.NoError(t, err)

	assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
		{"work/org/repo1", "git", "checkout", "-b", testsupport.Pwd()},
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo2"},
		{"work/org/repo2", "git", "checkout", "-b", testsupport.Pwd()},
	})
}

func TestItClonesReposInMultipleOrgs(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaignDirectory("orgA/repo1", "orgB/repo2")

	_, err := runCommand()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/orgA", "gh", "repo", "fork", "--clone=true", "orgA/repo1"},
		{"work/orgA/repo1", "git", "checkout", "-b", testsupport.Pwd()},
		{"work/orgB", "gh", "repo", "fork", "--clone=true", "orgB/repo2"},
		{"work/orgB/repo2", "git", "checkout", "-b", testsupport.Pwd()},
	})
}

func TestItClonesReposFromOtherHosts(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaignDirectory("mygitserver.com/orgA/repo1", "orgB/repo2")

	_, err := runCommand()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/orgA", "gh", "repo", "fork", "--clone=true", "mygitserver.com/orgA/repo1"},
		{"work/orgA/repo1", "git", "checkout", "-b", testsupport.Pwd()},
		{"work/orgB", "gh", "repo", "fork", "--clone=true", "orgB/repo2"},
		{"work/orgB/repo2", "git", "checkout", "-b", testsupport.Pwd()},
	})
}

func TestItSkipsCloningIfAWorkingCopyAlreadyExists(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	exec = fakeExecutor

	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")
	_ = os.MkdirAll(path.Join("work", "org", "repo1"), os.ModeDir|0755)

	out, err := runCommand()
	assert.NoError(t, err)
	assert.Contains(t, out, "Not cloning org/repo1")

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo2"},
		{"work/org/repo2", "git", "checkout", "-b", testsupport.Pwd()},
	})
}

func runCommand() (string, error) {
	cmd := CreateCloneCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()

	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
