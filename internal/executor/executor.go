package executor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
)

type Executor interface {
	Execute(output io.Writer, workingDir string, name string, args ...string) error
	ExecuteAndCapture(output io.Writer, workingDir string, name string, args ...string) (string, error)
}

type RealExecutor struct {
}

func (e *RealExecutor) Execute(output io.Writer, workingDir string, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = workingDir
	tailer(output)(command.StdoutPipe())
	tailer(output)(command.StderrPipe())

	_, err := fmt.Fprintln(output, "Executing:", name, summarizedArgs(args))
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	err = command.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (e *RealExecutor) ExecuteAndCapture(output io.Writer, workingDir string, name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	command.Dir = workingDir

	_, err := fmt.Fprintln(output, "Executing:", name, summarizedArgs(args))
	if err != nil {
		return "", err
	}

	commandOutput, err := command.Output()
	if err != nil {
		if exitErr, _ := err.(*exec.ExitError); exitErr != nil {
			stdErr := string(exitErr.Stderr)
			return stdErr, fmt.Errorf("Error: %w. Stderr: %s", exitErr, stdErr)
		}
		return string(commandOutput), err
	}

	return string(commandOutput), nil
}

func summarizedArgs(args []string) []string {
	result := []string{}
	for _, arg := range args {
		if len(arg) > 30 {
			result = append(result, "...")
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

func tailer(output io.Writer) func(io.ReadCloser, error) {
	return func(pipe io.ReadCloser, err error) {
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(pipe)
		go func() {
			for scanner.Scan() {
				_, err := fmt.Fprintf(output, "    %s\n", scanner.Text())
				if err != nil {
					return
				}
			}
		}()
	}
}
