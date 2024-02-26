package executor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutorExecuteVerbose(t *testing.T) {
	localExecutor := NewRealExecutor()
	outputBytes := bytes.NewBuffer([]byte{})

	err := localExecutor.Execute(outputBytes, ".", "echo", "Test1234")
	assert.NoError(t, err)

	output := outputBytes.String()

	assert.Contains(t, output, "Test1234")
	assert.Contains(t, output, "Executing: echo [Test1234]")
}

func TestExecutorExecuteSilent(t *testing.T) {
	localExecutor := NewRealExecutor()
	localExecutor.SetVerbose(false)

	outputBytes := bytes.NewBuffer([]byte{})

	err := localExecutor.Execute(outputBytes, ".", "echo", "Test1234")
	assert.NoError(t, err)

	output := outputBytes.String()
	assert.Contains(t, output, "Test1234")
	assert.NotContains(t, output, "Executing:")
}

func TestExecutorExecuteThrowsError(t *testing.T) {
	localExecutor := NewRealExecutor()

	outputBytes := bytes.NewBuffer([]byte{})

	err := localExecutor.Execute(outputBytes, ".", "fakecommand", "should", "error")
	assert.Error(t, err)

	output := outputBytes.String()
	assert.Contains(t, output, "Executing: fakecommand [should error] in .")
}

func TestExecutorExecuteAndCaptureVerbose(t *testing.T) {
	localExecutor := NewRealExecutor()
	commandOutput := bytes.NewBuffer([]byte{})

	output, err := localExecutor.ExecuteAndCapture(commandOutput, ".", "echo", "Test1234")
	assert.NoError(t, err)

	assert.Equal(t, output, "Test1234\n")
	assert.Contains(t, commandOutput.String(), "Executing: echo [Test1234]")
}

func TestExecutorExecuteAndCaptureSilent(t *testing.T) {
	localExecutor := NewRealExecutor()
	localExecutor.SetVerbose(false)
	commandOutput := bytes.NewBuffer([]byte{})

	output, err := localExecutor.ExecuteAndCapture(commandOutput, ".", "echo", "Test1234")
	assert.NoError(t, err)

	assert.Equal(t, output, "Test1234\n")
	assert.Empty(t, commandOutput.String())
}

func TestExecutorExecuteAndCaptureThrowsError(t *testing.T) {
	localExecutor := NewRealExecutor()
	commandOutput := bytes.NewBuffer([]byte{})

	output, err := localExecutor.ExecuteAndCapture(commandOutput, ".", "fakecommand", "does", "not", "exist")
	assert.Error(t, err)

	assert.Empty(t, output, "Test1234\n")
	assert.Contains(t, commandOutput.String(), "Executing: fakecommand [does not exist] in .")
}

func TestSummarizedArgs(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    []string
		Expected []string
	}{
		{
			TestName: "empty",
			Input:    []string{},
			Expected: []string{},
		},
		{
			TestName: "single",
			Input:    []string{"a"},
			Expected: []string{"a"},
		},
		{
			TestName: "multiple",
			Input:    []string{"a", "b", "c"},
			Expected: []string{"a", "b", "c"},
		},
		{
			TestName: "One of the arg is > 30 chars",
			Input:    []string{"a", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "c"},
			Expected: []string{"a", "...", "c"},
		},
		{
			TestName: "All Five args are long",
			Input: []string{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				"ccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				"ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
			},
			Expected: []string{"...", "...", "...", "...", "..."},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			assert.Equal(t, testCase.Expected, summarizedArgs(testCase.Input))
		})
	}
}
