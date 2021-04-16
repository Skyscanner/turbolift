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
}

type RealExecutor struct {
}

func (e *RealExecutor) Execute(output io.Writer, workingDir string, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = workingDir
	tailer(output)(command.StdoutPipe())
	tailer(output)(command.StderrPipe())

	_, err := fmt.Fprintln(output, "Executing:", name, args)
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
