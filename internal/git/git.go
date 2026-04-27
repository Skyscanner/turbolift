/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * https://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package git

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/skyscanner/turbolift/internal/executor"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type Git interface {
	Checkout(output io.Writer, workingDir string, branch string) error
	Push(stdout io.Writer, workingDir string, remote string, branchName string) error
	Commit(output io.Writer, workingDir string, message string) error
	IsRepoChanged(output io.Writer, workingDir string) (bool, error)
	Pull(output io.Writer, workingDir string, remote string, branchName string) error
	// GetCurrentBranchName returns the name of the currently checked-out branch
	// in workingDir. Used by `clone --from-prs` to capture a PR's branch after
	// `gh pr checkout` has fetched and checked it out — the PR branch name may
	// not match the campaign name, and it's simpler to read HEAD than to parse
	// `gh` output.
	GetCurrentBranchName(output io.Writer, workingDir string) (string, error)
}

type RealGit struct{}

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
	var localExecutor executor.Executor = executor.NewRealExecutor()
	localExecutor.SetVerbose(false)
	shellCommand := os.Getenv("SHELL")
	if shellCommand == "" {
		shellCommand = "sh"
	}
	shellArgs := []string{"-c", "git status --porcelain=v1 | wc -l | tr -d '[:space:]'"}
	commandOutput, err := localExecutor.ExecuteAndCapture(output, workingDir, shellCommand, shellArgs...)
	if err != nil {
		return false, err
	}

	diffSize, err := strconv.Atoi(commandOutput)
	if err != nil {
		return false, err
	}

	return diffSize > 0, nil
}

func (r *RealGit) Pull(output io.Writer, workingDir string, remote string, branchName string) error {
	return execInstance.Execute(output, workingDir, "git", "pull", "--ff-only", remote, branchName)
}

func (r *RealGit) GetCurrentBranchName(output io.Writer, workingDir string) (string, error) {
	// --abbrev-ref HEAD yields the short branch name if we're on a branch, or
	// "HEAD" if detached. Detached state shouldn't happen in the clone
	// --from-prs flow (gh pr checkout lands us on a branch) — but if it does,
	// the caller sees "HEAD" and can decide what to do rather than silently
	// writing "HEAD" into repos.txt.
	name, err := execInstance.ExecuteAndCapture(output, workingDir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(name), nil
}

func NewRealGit() *RealGit {
	return &RealGit{}
}
