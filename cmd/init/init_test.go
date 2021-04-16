package init

import (
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestAllFilesAreCreated(t *testing.T) {
	testsupport.CreateAndEnterTempDirectory()
	runCommand()

	assert.DirExistsf(t, "foo", "campaign directory should have been created")
	assert.DirExistsf(t, "foo/work", "work directory should have been created")
	assert.FileExists(t, "foo/.gitignore", "a .gitignore file should have been created")
	assert.FileExists(t, "foo/.turbolift", "a .turbolift file should have been created")
	assert.FileExists(t, "foo/README.md", "a README.md file should have been created")
	assert.FileExists(t, "foo/repos.txt", "a repos.txt file should have been created")
}

func TestTemplatedFilesHaveExpectedContent(t *testing.T) {
	testsupport.CreateAndEnterTempDirectory()
	runCommand()

	readmeContents, err := ioutil.ReadFile("foo/README.md")
	if err != nil {
		panic(err)
	}

	// Don't be too specific about expected content, to avoid test fragility
	assert.Contains(t, string(readmeContents), "foo")
}

func runCommand() {
	cmd := NewInitCmd()
	cmd.SetArgs([]string{"--name", "foo"})
	err := cmd.Execute()

	if err != nil {
		panic(err)
	}
}
