package clone

import (
	"bytes"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

var tests = map[string]func(t *testing.T){
	"it aborts if repos.txt not found": func(t *testing.T) {
		fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
		exec = fakeExecutor

		out, err := runCommand()
		assert.NoError(t, err)
		assert.Contains(t, out, "Error when reading campaign directory")

		fakeExecutor.AssertCalledWith(t, [][]string{})
	},
	"it logs clone errors but continues to try all": func(t *testing.T) {
		fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
		exec = fakeExecutor

		prepareTempCampaignDirectory("org/repo1", "org/repo2")

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
	},
	"it clones the repos found in repos.txt": func(t *testing.T) {
		fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
		exec = fakeExecutor

		prepareTempCampaignDirectory("org/repo1", "org/repo2")

		out, err := runCommand()
		assert.NoError(t, err)

		assert.Contains(t, out, "turbolift clone completed (2 repos cloned, 0 repos skipped)")

		fakeExecutor.AssertCalledWith(t, [][]string{
			{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
			{"work/org/repo1", "git", "checkout", "-b", pwd()},
			{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo2"},
			{"work/org/repo2", "git", "checkout", "-b", pwd()},
		})
	},
	"it clones repos in multiple orgs": func(t *testing.T) {
		fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
		exec = fakeExecutor

		prepareTempCampaignDirectory("orgA/repo1", "orgB/repo2")

		_, err := runCommand()
		assert.NoError(t, err)

		fakeExecutor.AssertCalledWith(t, [][]string{
			{"work/orgA", "gh", "repo", "fork", "--clone=true", "orgA/repo1"},
			{"work/orgA/repo1", "git", "checkout", "-b", pwd()},
			{"work/orgB", "gh", "repo", "fork", "--clone=true", "orgB/repo2"},
			{"work/orgB/repo2", "git", "checkout", "-b", pwd()},
		})
	},
	"it clones repos from other hosts": func(t *testing.T) {
		fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
		exec = fakeExecutor

		prepareTempCampaignDirectory("mygitserver.com/orgA/repo1", "orgB/repo2")

		_, err := runCommand()
		assert.NoError(t, err)

		fakeExecutor.AssertCalledWith(t, [][]string{
			{"work/orgA", "gh", "repo", "fork", "--clone=true", "mygitserver.com/orgA/repo1"},
			{"work/orgA/repo1", "git", "checkout", "-b", pwd()},
			{"work/orgB", "gh", "repo", "fork", "--clone=true", "orgB/repo2"},
			{"work/orgB/repo2", "git", "checkout", "-b", pwd()},
		})
	},
	"it skips cloning if a working copy already exists": func(t *testing.T) {
		fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
		exec = fakeExecutor

		prepareTempCampaignDirectory("org/repo1", "org/repo2")
		_ = os.MkdirAll(path.Join("work", "org", "repo1"), os.ModeDir|0755)

		out, err := runCommand()
		assert.NoError(t, err)
		assert.Contains(t, out, "Not cloning org/repo1")

		fakeExecutor.AssertCalledWith(t, [][]string{
			{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo2"},
			{"work/org/repo2", "git", "checkout", "-b", pwd()},
		})
	},
}

func pwd() string {
	dir, _ := os.Getwd()
	return path.Base(dir)
}

func setup() {
	tempDir, _ := ioutil.TempDir("", "turbolift-test-*")
	err := os.Chdir(tempDir)

	if err != nil {
		panic(err)
	}
}

func prepareTempCampaignDirectory(repos ...string) {
	delimitedList := strings.Join(repos, "\n")
	err := ioutil.WriteFile("repos.txt", []byte(delimitedList), os.ModePerm|0644)
	if err != nil {
		panic(err)
	}
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

func TestCloneCmd(t *testing.T) {
	for name, fn := range tests {
		setup()
		t.Run(name, fn)
	}
}
