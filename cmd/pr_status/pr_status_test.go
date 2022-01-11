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

package pr_status

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func init() {
	// disable output colouring so that strings we want to do 'Contains' checks on do not have ANSI escape sequences in IDEs
	_ = os.Setenv("NO_COLOR", "1")
}

func TestItLogsSummaryInformation(t *testing.T) {
	prepareFakeResponses()

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repo3")

	out, err := runCommand(false)
	assert.NoError(t, err)
	assert.Contains(t, out, "Checking PR status for org/repo1")
	assert.Contains(t, out, "Checking PR status for org/repo2")
	assert.Contains(t, out, "turbolift pr-status completed")
	assert.Regexp(t, "Open\\s+1", out)
	assert.Regexp(t, "Merged\\s+1", out)
	assert.Regexp(t, "Closed\\s+1", out)

	assert.Regexp(t, "Reactions: 👍\\s+4\\s+👎\\s+3\\s+🚀\\s+1", out)

	// Shouldn't show 'list' detailed info
	assert.NotRegexp(t, "org/repo1\\s+OPEN", out)
	assert.NotRegexp(t, "org/repo2\\s+MERGED", out)
}

func TestItLogsDetailedInformation(t *testing.T) {
	prepareFakeResponses()

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repo3")

	out, err := runCommand(true)
	assert.NoError(t, err)
	// Should still show summary info
	assert.Regexp(t, "Open\\s+1", out)
	assert.Regexp(t, "👍\\s+4", out)

	assert.Regexp(t, "org/repo1\\s+OPEN\\s+REVIEW_REQUIRED", out)
	assert.Regexp(t, "org/repo2\\s+MERGED\\s+APPROVED", out)
	assert.Regexp(t, "org/repo3\\s+CLOSED", out)
}

func TestItSkipsUnclonedRepos(t *testing.T) {
	prepareFakeResponses()

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2")
	_ = os.Remove("work/org/repo2")

	out, err := runCommand(true)
	assert.NoError(t, err)
	// Should still show summary info
	assert.Regexp(t, "Open\\s+1", out)
	assert.Regexp(t, "Merged\\s+0", out)
	assert.Regexp(t, "Skipped\\s+1", out)
	assert.Regexp(t, "Not Found\\s+0", out)

	assert.Regexp(t, "org/repo1\\s+OPEN", out)
	assert.NotRegexp(t, "org/repo2\\s+MERGED", out)
}

func TestItSkipsErroringRepos(t *testing.T) {
	prepareFakeResponses()

	testsupport.PrepareTempCampaign(true, "org/repo1", "org/repo2", "org/repoWithError")

	out, err := runCommand(true)
	assert.NoError(t, err)
	// Should still show summary info
	assert.Regexp(t, "Open\\s+1", out)
	assert.Regexp(t, "Merged\\s+1", out)
	assert.Regexp(t, "Skipped\\s+0", out)
	assert.Regexp(t, "Not Found\\s+1", out)

	assert.Regexp(t, "org/repo1\\s+OPEN", out)
}

func runCommand(showList bool) (string, error) {
	cmd := NewPrStatusCmd()
	list = showList
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	err := cmd.Execute()

	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}

func prepareFakeResponses() {
	dummyData := map[string]*github.PrStatus{
		"work/org/repo1": {
			State: "OPEN",
			ReactionGroups: []github.ReactionGroup{
				{
					Content: "THUMBS_UP",
					Users: github.ReactionGroupUsers{
						TotalCount: 3,
					},
				},
				{
					Content: "ROCKET",
					Users: github.ReactionGroupUsers{
						TotalCount: 1,
					},
				},
			},
			ReviewDecision: "REVIEW_REQUIRED",
		},
		"work/org/repo2": {
			State: "MERGED",
			ReactionGroups: []github.ReactionGroup{
				{
					Content: "THUMBS_UP",
					Users: github.ReactionGroupUsers{
						TotalCount: 1,
					},
				},
			},
			ReviewDecision: "APPROVED",
		},
		"work/org/repo3": {
			State: "CLOSED",
			ReactionGroups: []github.ReactionGroup{
				{
					Content: "THUMBS_DOWN",
					Users: github.ReactionGroupUsers{
						TotalCount: 3,
					},
				},
			},
		},
	}
	fakeGitHub := github.NewFakeGitHub(nil, func(output io.Writer, workingDir string) (interface{}, error) {
		if workingDir == "work/org/repoWithError" {
			return nil, errors.New("Synthetic error")
		} else {
			return dummyData[workingDir], nil
		}
	})
	gh = fakeGitHub
}
