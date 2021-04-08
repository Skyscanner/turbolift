package clone

import (
	"bytes"
	"github.com/skyscanner/turbolift/internal/executor"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
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

		fakeExecutor.AssertCalledWith(t, [][]string{
			{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
			{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo2"},
		})
	},
	"it clones the repos found in repos.txt": func(t *testing.T) {
		fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
		exec = fakeExecutor

		prepareTempCampaignDirectory("org/repo1", "org/repo2")

		_, err := runCommand()
		assert.NoError(t, err)

		fakeExecutor.AssertCalledWith(t, [][]string{
			{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo1"},
			{"work/org", "gh", "repo", "fork", "--clone=true", "org/repo2"},
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
			{"work/orgB", "gh", "repo", "fork", "--clone=true", "orgB/repo2"},
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
			{"work/orgB", "gh", "repo", "fork", "--clone=true", "orgB/repo2"},
		})
	},
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
