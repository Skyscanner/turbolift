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
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type FakeGitHub struct {
	handler func(output io.Writer, workingDir string, fullRepoName string) (bool, error)
	calls   [][]string
}

func (f *FakeGitHub) CreatePullRequest(output io.Writer, workingDir string, metadata PullRequest) (didCreate bool, err error) {
	f.calls = append(f.calls, []string{workingDir, metadata.Title})
	return f.handler(output, workingDir, "")
}

func (f *FakeGitHub) ForkAndClone(output io.Writer, workingDir string, fullRepoName string) error {
	f.calls = append(f.calls, []string{workingDir, fullRepoName})
	_, err := f.handler(output, workingDir, fullRepoName)
	return err
}

func (f *FakeGitHub) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGitHub(h func(output io.Writer, workingDir string, fullRepoName string) (bool, error)) *FakeGitHub {
	return &FakeGitHub{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) (bool, error) {
		return true, nil
	})
}

func NewAlwaysFailsFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) (bool, error) {
		return false, errors.New("synthetic error")
	})
}

func NewAlwaysReturnsFalseFakeGitHub() *FakeGitHub {
	return NewFakeGitHub(func(output io.Writer, workingDir string, fullRepoName string) (bool, error) {
		return false, nil
	})
}
