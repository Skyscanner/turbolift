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

package git

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io"
	"os"
	"strconv"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type Git interface {
	Checkout(output io.Writer, workingDir string, branch string) error
	Push(stdout io.Writer, workingDir string, remote string, branchName string) error
	Commit(output io.Writer, workingDir string, message string) error
	IsRepoChanged(output io.Writer, workingDir string) (bool, error)
}

type RealGit struct {
}

func (r *RealGit) Checkout(output io.Writer, workingDir string, branchName string) error {
	return execInstance.Execute(output, workingDir, "git", "checkout", "-b", branchName)
}

func (r *RealGit) Push(output io.Writer, workingDir string, remote string, branchName string) error {
	return execInstance.Execute(output, workingDir, "git", "push", "-u", remote, branchName)
}

func (r *RealGit) Commit(output io.Writer, workingDir string, message string) error {
	return execInstance.Execute(output, workingDir, "git", "commit", "--all", "--message", message)
}

func (r *RealGit) IsRepoChanged(output io.Writer, workingDir string) (bool, error) {
	shellCommand := os.Getenv("SHELL")
	if shellCommand == "" {
		shellCommand = "sh"
	}
	shellArgs := []string{"-c", "git status --porcelain=v1 | wc -l | tr -d '[:space:]'"}
	commandOutput, err := execInstance.ExecuteAndCapture(output, workingDir, shellCommand, shellArgs...)

	if err != nil {
		return false, err
	}

	diffSize, err := strconv.Atoi(commandOutput)
	if err != nil {
		return false, err
	}

	return diffSize > 0, nil
}

func NewRealGit() *RealGit {
	return &RealGit{}
}
