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
 */

package campaign

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePRRef_GithubComURL(t *testing.T) {
	pr, err := ParsePRRef("https://github.com/Skyscanner/turbolift/pull/215")
	assert.NoError(t, err)
	assert.Equal(t, PRRef{Host: "", OrgName: "Skyscanner", RepoName: "turbolift", Number: 215}, pr)
}

func TestParsePRRef_GHEURL(t *testing.T) {
	pr, err := ParsePRRef("https://github.my-ghe.example/org/repo/pull/42")
	assert.NoError(t, err)
	assert.Equal(t, PRRef{Host: "github.my-ghe.example", OrgName: "org", RepoName: "repo", Number: 42}, pr)
}

func TestParsePRRef_Shorthand(t *testing.T) {
	pr, err := ParsePRRef("Skyscanner/turbolift#216")
	assert.NoError(t, err)
	assert.Equal(t, PRRef{OrgName: "Skyscanner", RepoName: "turbolift", Number: 216}, pr)
}

func TestParsePRRef_Malformed(t *testing.T) {
	_, err := ParsePRRef("not a pr ref at all")
	assert.Error(t, err)
}

func TestParsePRRef_URLMissingNumber(t *testing.T) {
	_, err := ParsePRRef("https://github.com/org/repo/pull/")
	assert.Error(t, err)
}

func TestReadPRsFile_HappyPath(t *testing.T) {
	tmp, err := os.CreateTemp("", "prs-*.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, _ = tmp.WriteString("# a full-line comment\n\nhttps://github.com/org/repo1/pull/1\norg/repo2#2\n")
	_ = tmp.Close()

	prs, err := ReadPRsFile(tmp.Name())
	assert.NoError(t, err)
	assert.Equal(t, []PRRef{
		{OrgName: "org", RepoName: "repo1", Number: 1},
		{OrgName: "org", RepoName: "repo2", Number: 2},
	}, prs)
}

func TestReadPRsFile_DuplicatePRDeduped(t *testing.T) {
	tmp, err := os.CreateTemp("", "prs-*.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, _ = tmp.WriteString("org/repo1#1\norg/repo1#1\n")
	_ = tmp.Close()

	prs, err := ReadPRsFile(tmp.Name())
	assert.NoError(t, err)
	assert.Len(t, prs, 1)
}

func TestReadPRsFile_SameRepoDifferentPRsErrors(t *testing.T) {
	tmp, err := os.CreateTemp("", "prs-*.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, _ = tmp.WriteString("org/repo1#1\norg/repo1#2\n")
	_ = tmp.Close()

	_, err = ReadPRsFile(tmp.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple PRs")
}

func TestReadPRsFile_MalformedLine(t *testing.T) {
	tmp, err := os.CreateTemp("", "prs-*.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, _ = tmp.WriteString("not a pr ref\n")
	_ = tmp.Close()

	_, err = ReadPRsFile(tmp.Name())
	assert.Error(t, err)
}
