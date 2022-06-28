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

type GitHub interface {
	ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error
	Clone(output io.Writer, workingDir string, fullRepoName string) error
	CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error)
	ClosePullRequest(output io.Writer, workingDir string, branchName string) error
	GetPR(output io.Writer, workingDir string, branchName string) (*PrStatus, error)
}

type RealGitHub struct{}

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

func (r *RealGitHub) ClosePullRequest(output io.Writer, workingDir string, branchName string) error {
	pr, err := r.GetPR(output, workingDir, branchName)
	if err != nil {
		return err
	}

	return execInstance.Execute(output, workingDir, "gh", "pr", "close", fmt.Sprint(pr.Number))
}

// the following is used internally to retrieve PRs from a given repository
// using `gh pr status`

type PrStatusResponse struct { // https://github.com/cli/cli/blob/4b415f80d79e57eda48bb67b30cfb53d18b7cba7/pkg/cmd/pr/status/status.go#L114-L118
	CurrentBranch *PrStatus   `json:"currentBranch"`
	CreatedBy     []*PrStatus `json:"createdBy"`
	NeedsReview   []*PrStatus `json:"needsReview"`
}

type PrStatus struct {
	Closed         bool            `json:"closed"`
	HeadRefName    string          `json:"headRefName"`
	Mergeable      string          `json:"mergeable"`
	Number         int             `json:"number"`
	ReactionGroups []ReactionGroup `json:"reactionGroups"`
	ReviewDecision string          `json:"reviewDecision"`
	State          string          `json:"state"`
	Title          string          `json:"title"`
	Url            string          `json:"url"`
}

type ReactionGroupUsers struct {
	TotalCount int
}

type ReactionGroup struct {
	Content string
	Users   ReactionGroupUsers
}

// GetPR is a helper function to retrieve the PR associated with the branch Name
type NoPRFoundError struct {
	Path       string
	BranchName string
}

func (e *NoPRFoundError) Error() string {
	return fmt.Sprintf("no PR found for %s and branch %s", e.Path, e.BranchName)
}

func (r *RealGitHub) GetPR(output io.Writer, workingDir string, branchName string) (*PrStatus, error) {
	s, err := execInstance.ExecuteAndCapture(output, workingDir, "gh", "pr", "status", "--json", "closed,headRefName,mergeable,number,reactionGroups,reviewDecision,state,title,url")
	if err != nil {
		return nil, err
	}

	var prr PrStatusResponse
	if err := json.Unmarshal([]byte(s), &prr); err != nil {
		return nil, fmt.Errorf("unable to unmarshall the Pr Status Output: %w", err)
	}

	// if the user has write permissions on the repo,
	// the PR should be under _CurrentBranch_.
	if prr.CurrentBranch != nil && !prr.CurrentBranch.Closed {
		return prr.CurrentBranch, nil
	}

	// If not, then it's a forked PR. The headRefName is as such: `username:branchName`
	for _, pr := range prr.CreatedBy {
		if strings.HasSuffix(pr.HeadRefName, branchName) && !pr.Closed {
			return pr, nil
		}
	}

	return nil, &NoPRFoundError{Path: workingDir, BranchName: branchName}
}

func NewRealGitHub() *RealGitHub {
	return &RealGitHub{}
}
