package logging

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/skyscanner/turbolift/cmd/flags"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/spf13/cobra"
	"io"
	"time"
)

// Logger is a facade for CLI logging.
type Logger struct {
	writer  io.Writer
	verbose bool
}

// NewLogger creates a Logger associated with a particular *cobra.Command instance.
// Logs will be delivered to the command's stdout writer.
func NewLogger(c *cobra.Command) *Logger {
	return &Logger{
		writer:  c.OutOrStdout(),
		verbose: flags.Verbose,
	}
}

func (log *Logger) Printf(s string, args ...interface{}) {
	_, _ = fmt.Fprintf(log.writer, s, args...)
	_, _ = fmt.Fprintln(log.writer)
}

func (log *Logger) Println(s ...interface{}) {
	_, _ = fmt.Fprintln(log.writer, s...)
}

func (log *Logger) Successf(format string, args ...interface{}) {
	log.Printf(colors.Green(format), args...)
}

func (log *Logger) Warnf(format string, args ...interface{}) {
	log.Printf(colors.Yellow(format), args...)
}

// StartActivity creates and starts an *Activity with an associated spinner.
// Only once Activity should be active at any given time, and the Activity should be completed before any other logging
// is performed using this Logger.
func (log *Logger) StartActivity(format string, args ...interface{}) *Activity {
	name := fmt.Sprintf(format, args...)
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond) // Build our new spinner
	s.Suffix = fmt.Sprintf("  %s", name)
	s.Writer = log.writer
	s.HideCursor = true
	s.Start()

	return &Activity{
		name:    name,
		logs:    []string{},
		spinner: s,
		writer:  log.writer,
		verbose: log.verbose,
	}
}
