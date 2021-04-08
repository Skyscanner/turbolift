package executor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
)

type Executor interface {
	Execute(workingDir string, name string, args ...string) error
}

type RealExecutor struct {
}

func (e *RealExecutor) Execute(workingDir string, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = workingDir
	tailer()(command.StdoutPipe())
	tailer()(command.StderrPipe())

	fmt.Println("Executing:", name, args)
	if err := command.Start(); err != nil {
		return err
	}

	err := command.Wait()
	if err != nil {
		return err
	}

	return nil
}

func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

func tailer() func(io.ReadCloser, error) {
	return func(pipe io.ReadCloser, err error) {
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(pipe)
		go func() {
			for scanner.Scan() {
				fmt.Printf("    %s\n", scanner.Text())
			}
		}()
	}
}
