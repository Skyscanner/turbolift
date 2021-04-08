package clone

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io/ioutil"
	"os"
	"testing"
)

var tests = map[string]func(t *testing.T){
	"it is run": func(t *testing.T) {
		exec = executor.NewFakeExecutor(func(s string, s2 ...string) error {
			return nil
		})

		runCommand()
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
	cmd := CreateCloneCmd()
	err := cmd.Execute()

	if err != nil {
		panic(err)
	}
}

func TestCloneCmd(t *testing.T) {
	for name, fn := range tests {
		setup()
		t.Run(name, fn)
	}
}
