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
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeGitHub struct {
	handler          func(output io.Writer, command Command, workingDir string, fullRepoName string) (bool, error)
	returningHandler func(output io.Writer, workingDir string) (interface{}, error)
	calls            [][]string
}

func (f *FakeGitHub) CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error) {
	f.calls = append(f.calls, []string{workingDir, metadata.Title})
	return f.handler(output, CREATE_PULL_REQUEST, workingDir, "")
}

func (f *FakeGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	f.calls = append(f.calls, []string{workingDir, fullRepoName})
	_, err := f.handler(output, FORK_AND_CLONE, workingDir, fullRepoName)
	return err
}

func (f *FakeGitHub) Clone(output io.Writer, workingDir string, fullRepoName string) error {
	f.calls = append(f.calls, []string{workingDir, fullRepoName})
	_, err := f.handler(output, CLONE, workingDir, fullRepoName)
	return err
}

func (f *FakeGitHub) ClosePullRequest(output io.Writer, workingDir string, branchName string) error {
	// TODO: handle this differently; branchName here is replacing fullRepoName
	// This is OK for now because fullRepoName is used nowhere in the github mocks
	f.calls = append(f.calls, []string{workingDir, branchName})
	_, err := f.handler(output, CLOSE_PULL_REQUEST, workingDir, branchName)
	return err
}

func (f *FakeGitHub) GetPR(output io.Writer, workingDir string, _ string) (*PrStatus, error) {
	f.calls = append(f.calls, []string{workingDir})
	result, err := f.returningHandler(output, workingDir)
	if result == nil {
		return nil, err
	}
	return result.(*PrStatus), err
}

func (f *FakeGitHub) GetDefaultBranchName(output io.Writer, workingDir string, fullRepoName string) (string, error) {
	f.calls = append(f.calls, []string{workingDir, fullRepoName})
	_, err := f.handler(output, GET_DEFAULT_BRANCH_NAME, workingDir, fullRepoName)
	return "main", err
}

func (f *FakeGitHub) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGitHub(h func(output io.Writer, command Command, workingDir string, fullRepoName string) (bool, error), r func(output io.Writer, workingDir string) (interface{}, error)) *FakeGitHub {
	return &FakeGitHub{
		handler:          h,
		returningHandler: r,
		calls:            [][]string{},
	}
}

func NewAlwaysSucceedsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, command Command, workingDir string, fullRepoName string) (bool, error) {
		return true, nil
	}, func(output io.Writer, workingDir string) (interface{}, error) {
		return PrStatus{}, nil
	})
}

func NewAlwaysFailsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, command Command, workingDir string, fullRepoName string) (bool, error) {
		return false, errors.New("synthetic error")
	}, func(output io.Writer, workingDir string) (interface{}, error) {
		return nil, errors.New("synthetic error")
	})
}

func NewAlwaysThrowNoPRFound() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, command Command, workingDir string, branchName string) (bool, error) {
		return false, &NoPRFoundError{Path: workingDir, BranchName: branchName}
	}, func(output io.Writer, workingDir string) (interface{}, error) {
		panic("should not be invoked")
	})
}

func NewAlwaysReturnsFalseFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, command Command, workingDir string, fullRepoName string) (bool, error) {
		return false, nil
	}, func(output io.Writer, workingDir string) (interface{}, error) {
		return PrStatus{}, nil
	})
}

func NewAlwaysFailsOnGetDefaultBranchFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, command Command, workingDir string, fullRepoName string) (bool, error) {
		if command == GET_DEFAULT_BRANCH_NAME {
			return false, errors.New("synthetic error")
		}
		return true, nil
	}, func(output io.Writer, workingDir string) (interface{}, error) {
		return PrStatus{}, nil
	})
}

type Command int

const (
	FORK_AND_CLONE          Command = 0
	CLONE                   Command = 1
	CREATE_PULL_REQUEST     Command = 2
	CLOSE_PULL_REQUEST      Command = 3
	GET_DEFAULT_BRANCH_NAME Command = 4
)
