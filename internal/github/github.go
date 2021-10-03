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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/skyscanner/turbolift/internal/executor"
)

var execInstance executor.Executor = executor.NewRealExecutor()

type PullRequest struct {
	Title          string
	Body           string
	UpstreamRepo   string
	IsDraft        bool
	ReviewDecision string
}

type ReactionGroupUsers struct {
	TotalCount int
}

type ReactionGroup struct {
	Content string
	Users   ReactionGroupUsers
}

type PrStatus struct {
	Mergeable      string
	ReviewDecision string
	State          string
	ReactionGroups []ReactionGroup
	Url            string
}

type GitHub interface {
	ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error
	Clone(output io.Writer, workingDir string, fullRepoName string) error
	CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error)
	GetPrStatus(output io.Writer, workingDir string) (*PrStatus, error)
}

type RealGitHub struct {
}

func (r *RealGitHub) CreatePullRequest(output io.Writer, workingDir string, pr PullRequest) (didCreate bool, err error) {
	gh_args := []string{
		"pr",
		"create",
		"--title",
		pr.Title,
		"--body",
		pr.Body,
		"--repo",
		pr.UpstreamRepo,
	}

	if pr.IsDraft {
		gh_args = append(gh_args, "--draft")
	}

	execOutput, err := execInstance.ExecuteAndCapture(output, workingDir, "gh", gh_args...)
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

func (r *RealGitHub) Clone(output io.Writer, workingDir string, fullRepoName string) error {
	return execInstance.Execute(output, workingDir, "gh", "repo", "clone", fullRepoName)
}

func (r *RealGitHub) GetPrStatus(output io.Writer, workingDir string) (*PrStatus, error) {
	s, err := execInstance.ExecuteAndCapture(output, workingDir, "gh", "pr", "view", "--json", "title,state,mergeable,statusCheckRollup,reviewDecision,reactionGroups,url")
	if err != nil {
		return nil, err
	}

	var status PrStatus
	if err := json.Unmarshal([]byte(s), &status); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON for PR status: %w", err)
	}
	return &status, nil
}

func NewRealGitHub() *RealGitHub {
	return &RealGitHub{}
}
