package clone

import (
	"github.com/skyscanner/turbolift/internal/fake_executor"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

var tests = map[string]func(t *testing.T){
	"it is run": func(t *testing.T) {
		execCommand = fake_executor.NewFakeExecCommand(0)
		defer func() { execCommand = exec.Command }()

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

func TestMain(m *testing.M) {
	fake_executor.SupportTestMain(m)
}
