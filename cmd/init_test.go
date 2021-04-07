package cmd

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var tests = map[string]func(t *testing.T){
	"all files are created": func(t *testing.T) {
		runCommand()

		assert.DirExistsf(t, "foo", "campaign directory should have been created")
		assert.DirExistsf(t, "foo/work", "work directory should have been created")
		assert.FileExists(t, "foo/.gitignore", "a .gitignore file should have been created")
		assert.FileExists(t, "foo/.turbolift", "a .turbolift file should have been created")
		assert.FileExists(t, "foo/README.md", "a README.md file should have been created")
		assert.FileExists(t, "foo/repos.txt", "a repos.txt file should have been created")
	},
	"templated files have expected content": func(t *testing.T) {
		runCommand()

		readmeContents, err := ioutil.ReadFile("foo/README.md")
		if err != nil {
			panic(err)
		}

		// Don't be too specific about expected content, to avoid test fragility
		assert.Contains(t, string(readmeContents), "foo")
	},
}

func setup() {
	tempDir, _ := ioutil.TempDir("", "turbolift-test-*")
	err := os.Chdir(tempDir)

	if err != nil {
		panic(err)
	}
}

func runCommand() {
	cmd := createInitCmd()
	cmd.SetArgs([]string{"--name", "foo"})
	err := cmd.Execute()

	if err != nil {
		panic(err)
	}
}

func TestInitCmd(t *testing.T) {
	for name, fn := range tests {
		setup()
		t.Run(name, fn)
	}
}
