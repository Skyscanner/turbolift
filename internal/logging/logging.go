/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
	prefixedFormat := fmt.Sprintf("✅ %s", format)
	log.Printf(colors.Green(prefixedFormat), args...)
}

func (log *Logger) Warnf(format string, args ...interface{}) {
	prefixedFormat := fmt.Sprintf("⚠️ %s", format)
	log.Printf(colors.Yellow(prefixedFormat), args...)
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
