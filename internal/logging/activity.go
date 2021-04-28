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
	"github.com/skyscanner/turbolift/internal/colors"
	"io"
	"strings"
)

// Activity is a buffered logger associated with an on-screen spinner.
// As well as being able to signal completion state (EndWithSuccess, EndWithWarning and EndWithFailure), logs can be
// buffered. Whether or not the logs are actually displayed depends on the completion state.
type Activity struct {
	name    string
	logs    []string
	spinner *spinner.Spinner
	writer  io.Writer
	verbose bool
}

func (a *Activity) Log(message string) {
	a.logs = append(a.logs, message)
}

func (a *Activity) Logf(format string, args ...interface{}) {
	a.Log(fmt.Sprintf(format, args...))
}

func (a *Activity) emitLogs(colourTransform func(...interface{}) string) {
	_, _ = fmt.Fprintln(a.writer)

	for _, log := range a.logs {
		_, _ = fmt.Fprint(a.writer, "     ")
		_, _ = fmt.Fprintln(a.writer, colourTransform(log))
	}
}

func (a *Activity) EndWithSuccess() {
	a.spinner.FinalMSG = fmt.Sprintf("✅ %s", a.name)
	a.spinner.Stop()
	_, _ = fmt.Fprintln(a.writer)

	if a.verbose {
		a.emitLogs(colors.White)
	}
}

func (a *Activity) EndWithSuccessAndEmitLogs() {
	a.spinner.FinalMSG = fmt.Sprintf("✅ %s", a.name)
	a.spinner.Stop()
	_, _ = fmt.Fprintln(a.writer)

	a.emitLogs(colors.White)
}

func (a *Activity) EndWithWarning(message interface{}) {
	a.spinner.FinalMSG = fmt.Sprintf(colors.Yellow("⚠️  %s: %s"), a.name, message)
	a.spinner.Stop()
	_, _ = fmt.Fprintln(a.writer)

	a.emitLogs(colors.Yellow)
}

func (a *Activity) EndWithWarningf(format string, args ...interface{}) {
	a.EndWithWarning(fmt.Sprintf(format, args...))
}

func (a *Activity) EndWithFailure(message interface{}) {
	a.spinner.FinalMSG = fmt.Sprintf(colors.Red("❌ %s: %s"), a.name, message)
	a.spinner.Stop()
	_, _ = fmt.Fprintln(a.writer)

	a.emitLogs(colors.Red)
}

func (a *Activity) EndWithFailuref(format string, args ...interface{}) {
	a.EndWithFailure(fmt.Sprintf(format, args...))
}

type logWriter struct {
	activity *Activity
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	line := string(p)
	trimmedLine := strings.TrimRight(line, "\n")

	l.activity.Log(trimmedLine)
	return len(p), nil
}

func (a *Activity) Writer() io.Writer {
	return &logWriter{
		activity: a,
	}
}
