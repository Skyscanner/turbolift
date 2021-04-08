package executor

import (
	"bufio"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os/exec"
)

type Executor interface {
	Execute(c *cobra.Command, workingDir string, name string, args ...string) error
}

type RealExecutor struct {
}

func (e *RealExecutor) Execute(c *cobra.Command, workingDir string, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = workingDir
	tailer(c)(command.StdoutPipe())
	tailer(c)(command.StderrPipe())

	c.Println("Executing:", name, args)
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

func tailer(c *cobra.Command) func(io.ReadCloser, error) {
	return func(pipe io.ReadCloser, err error) {
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(pipe)
		go func() {
			for scanner.Scan() {
				c.Printf("    %s\n", scanner.Text())
			}
		}()
	}
}
