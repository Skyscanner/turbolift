package clone

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os/exec"
)

var execCommand = exec.Command

func CreateCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: " ", // TODO
		Run:   run,
	}

	return cmd
}

func run(*cobra.Command, []string) {
	command := execCommand("gh", "repo", "clone", "mshell/mshell-tools")
	tailer("gh")(command.StdoutPipe())
	tailer("gh")(command.StderrPipe())

	if err := command.Start(); err != nil {
		log.Fatal(err)
	}

	err := command.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func tailer(label string) func(io.ReadCloser, error) {
	return func(pipe io.ReadCloser, err error) {
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(pipe)
		go func() {
			for scanner.Scan() {
				fmt.Printf("%s | ", label)
				fmt.Printf("%s\n", scanner.Text())
			}
		}()
	}
}
