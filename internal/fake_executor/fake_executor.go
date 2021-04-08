package fake_executor

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

// Creates a new exec.Cmd which will be diverted away from the original executable towards a fake implementation for
// testing.
//
// Usage (assuming execCommand is a variable normally pointing to exec.Command):
//     	execCommand = fake_executor.NewFakeExecCommand(0)
//		defer func() {execCommand = exec.Command}()
//      ... do something which uses execCommand ...
//
// Takes exitCode, a parameter indicating the exit code to be returned by 'the fake executable'
func NewFakeExecCommand(exitCode int) func(command string, args ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		c := exec.Command(os.Args[0])
		c.Env = os.Environ()
		c.Env = append(c.Env, "TEST_MAIN=fake_executor")
		c.Env = append(c.Env, fmt.Sprintf("TEST_FAKE_EXIT_CODE=%d", exitCode))
		return c
	}
}

// TestMain implementation for tests that use fake executor. Usage:
//		func TestMain(m *testing.M) {
//			fake_executor.SupportTestMain(m)
//		}
func SupportTestMain(m *testing.M) {
	switch os.Getenv("TEST_MAIN") {
	case "fake_executor":
		actAsFake()
	default:
		os.Exit(m.Run())
	}
}

// the fake implementation, invoked when the fake executor is in use
func actAsFake() {
	// Fake exit code
	fakeExitCode, err := strconv.Atoi(os.Getenv("TEST_FAKE_EXIT_CODE"))
	if err != nil {
		panic(err)
	}

	os.Exit(int(fakeExitCode))
}
