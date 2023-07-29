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
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestItReturnsErrorOnFailedCheckout(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runCheckoutAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "git", "checkout", "-b", "some_branch"},
	})
}

func TestItReturnsNilErrorOnSuccessfulCheckout(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runCheckoutAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org/repo1", "git", "checkout", "-b", "some_branch"},
	})
}

func TestItReturnsErrorOnFailedPull(t *testing.T) {
	fakeExecutor := executor.NewAlwaysFailsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runPullAndCaptureOutput()
	assert.Error(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org1/repo1", "git", "pull", "--ff-only", "upstream", "main"},
	})
}

func TestItReturnsNilErrorOnSuccessfulPull(t *testing.T) {
	fakeExecutor := executor.NewAlwaysSucceedsFakeExecutor()
	execInstance = fakeExecutor

	_, err := runPullAndCaptureOutput()
	assert.NoError(t, err)

	fakeExecutor.AssertCalledWith(t, [][]string{
		{"work/org1/repo1", "git", "pull", "--ff-only", "upstream", "main"},
	})
}

func runCheckoutAndCaptureOutput() (string, error) {
	sb := strings.Builder{}
	err := NewRealGit().Checkout(&sb, "work/org/repo1", "some_branch")

	if err != nil {
		return sb.String(), err
	}
	return sb.String(), nil
}

func runPullAndCaptureOutput() (string, error) {
	sb := strings.Builder{}
	err := NewRealGit().Pull(&sb, "work/org1/repo1", "upstream", "main")

	return sb.String(), err
}
