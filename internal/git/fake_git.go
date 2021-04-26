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
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type FakeGit struct {
	handler func(output io.Writer, call []string) (bool, error)
	calls   [][]string
}

func (f *FakeGit) Checkout(output io.Writer, workingDir string, branch string) error {
	call := []string{"checkout", workingDir, branch}
	f.calls = append(f.calls, call)
	_, err := f.handler(output, call)
	return err
}

func (f *FakeGit) Commit(output io.Writer, workingDir string, message string) error {
	call := []string{"commit", workingDir, message}
	f.calls = append(f.calls, call)
	_, err := f.handler(output, call)
	return err
}

func (f *FakeGit) IsRepoChanged(output io.Writer, workingDir string) (bool, error) {
	call := []string{"isRepoChanged", workingDir}
	f.calls = append(f.calls, call)
	result, err := f.handler(output, call)
	return result, err
}

func (f *FakeGit) Push(output io.Writer, workingDir string, _ string, branchName string) error {
	call := []string{"push", workingDir, branchName}
	f.calls = append(f.calls, call)
	_, err := f.handler(output, call)
	return err
}

func (f *FakeGit) AssertCalledWith(t *testing.T, expected [][]string) {
	assert.Equal(t, expected, f.calls)
}

func NewFakeGit(h func(io.Writer, []string) (bool, error)) *FakeGit {
	return &FakeGit{
		handler: h,
		calls:   [][]string{},
	}
}

func NewAlwaysSucceedsFakeGit() *FakeGit {
	return NewFakeGit(func(io.Writer, []string) (bool, error) {
		return true, nil
	})
}

func NewAlwaysFailsFakeGit() *FakeGit {
	return NewFakeGit(func(io.Writer, []string) (bool, error) {
		return false, errors.New("synthetic error")
	})
}
