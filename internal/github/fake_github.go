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

package github

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Command int

const (
	ForkAndClone Command = iota
	Clone
	CreatePullRequest
	ClosePullRequest
	GetDefaultBranchName
	UpdatePRDescription
	IsPushable
)

type FakeGitHub struct {
	handler          func(command Command, args []string) (bool, error)
	returningHandler func(workingDir string) (interface{}, error)
	calls            [][]string
}

func (f *FakeGitHub) CreatePullRequest(_ io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error) {
	labelValue := ""
	if metadata.ApplyLabels {
		labelValue = TurboliftLabel
	}
	args := []string{"create_pull_request", workingDir, metadata.Title, labelValue}
	f.calls = append(f.calls, args)
	return f.handler(CreatePullRequest, args)
}

func (f *FakeGitHub) ForkAndClone(_ io.Writer, workingDir string, fullRepoName string) error {
	args := []string{"fork_and_clone", workingDir, fullRepoName}
	f.calls = append(f.calls, args)
	_, err := f.handler(ForkAndClone, args)
	return err
}

func (f *FakeGitHub) Clone(_ io.Writer, workingDir string, fullRepoName string) error {
	args := []string{"clone", workingDir, fullRepoName}
	f.calls = append(f.calls, args)
	_, err := f.handler(Clone, args)
	return err
}

func (f *FakeGitHub) IsPushable(_ io.Writer, repo string) (bool, error) {
	args := []string{"user_can_push", repo}
	f.calls = append(f.calls, args)
	return f.handler(IsPushable, args)
}

func (f *FakeGitHub) ClosePullRequest(_ io.Writer, workingDir string, branchName string) error {
	args := []string{"close_pull_request", workingDir, branchName}
	f.calls = append(f.calls, args)
	_, err := f.handler(ClosePullRequest, args)
	return err
}

func (f *FakeGitHub) GetPR(_ io.Writer, workingDir string, _ string) (*PrStatus, error) {
	f.calls = append(f.calls, []string{"get_pr", workingDir})
	result, err := f.returningHandler(workingDir)
	if result == nil {
		return nil, err
	}
	return result.(*PrStatus), err
}

func (f *FakeGitHub) GetDefaultBranchName(_ io.Writer, workingDir string, fullRepoName string) (string, error) {
	args := []string{"get_default_branch", workingDir, fullRepoName}
	f.calls = append(f.calls, args)
	_, err := f.handler(GetDefaultBranchName, args)
	return "main", err
}

func (f *FakeGitHub) UpdatePRDescription(_ io.Writer, workingDir string, title string, body string) error {
	args := []string{"update_pr_description", workingDir, title, body}
	f.calls = append(f.calls, args)
	_, err := f.handler(UpdatePRDescription, args)
	return err
}

func (f *FakeGitHub) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGitHub(h func(command Command, args []string) (bool, error), r func(workingDir string) (interface{}, error)) *FakeGitHub {
	return &FakeGitHub{
		handler:          h,
		returningHandler: r,
		calls:            [][]string{},
	}
}

func NewAlwaysSucceedsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(command Command, args []string) (bool, error) {
		return true, nil
	}, func(workingDir string) (interface{}, error) {
		return PrStatus{}, nil
	})
}

func NewAlwaysFailsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(command Command, args []string) (bool, error) {
		return false, errors.New("synthetic error")
	}, func(workingDir string) (interface{}, error) {
		return nil, errors.New("synthetic error")
	})
}

func NewAlwaysThrowNoPRFound() *FakeGitHub {
	return NewFakeGitHub(func(command Command, args []string) (bool, error) {
		workingDir, branchName := args[1], args[2]
		return false, &NoPRFoundError{Path: workingDir, BranchName: branchName}
	}, func(workingDir string) (interface{}, error) {
		panic("should not be invoked")
	})
}

func NewAlwaysReturnsFalseFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(command Command, args []string) (bool, error) {
		return false, nil
	}, func(workingDir string) (interface{}, error) {
		return PrStatus{}, nil
	})
}

func NewAlwaysFailsOnGetDefaultBranchFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(command Command, args []string) (bool, error) {
		if command == GetDefaultBranchName {
			return false, errors.New("synthetic error")
		}
		return true, nil
	}, func(workingDir string) (interface{}, error) {
		return PrStatus{}, nil
	})
}
