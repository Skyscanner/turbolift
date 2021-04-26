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

package github

import (
	"github.com/skyscanner/turbolift/internal/executor"
	"io"
	"strings"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type PullRequest struct {
	Title        string
	Body         string
	UpstreamRepo string
}

type GitHub interface {
	ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error
	CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error)
}

type RealGitHub struct {
}

func (r *RealGitHub) CreatePullRequest(output io.Writer, workingDir string, pr PullRequest) (didCreate bool, err error) {
	execOutput, err := execInstance.ExecuteAndCapture(output, workingDir, "gh", "pr", "create", "--title", pr.Title, "--body", pr.Body, "--repo", pr.UpstreamRepo)

	if strings.Contains(execOutput, "GraphQL error: No commits between") {
		// no PR was created because there are no differences between remotes
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RealGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	return execInstance.Execute(output, workingDir, "gh", "repo", "fork", "--clone=true", fullRepoName)
}

func NewRealGitHub() *RealGitHub {
	return &RealGitHub{}
}
