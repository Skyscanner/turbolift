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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeReposFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "repos.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(b)
}

func TestUpsertBranchAnnotations_AppendsToEmptyTemplate(t *testing.T) {
	initial := "# List repositories to be operated upon in this file, one repo per line. e.g.\n# org/repo1\n# org/repo2\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{
		"org/repo1": "feature-x",
	})
	assert.NoError(t, err)

	expected := initial + "org/repo1 # branch=feature-x\n"
	assert.Equal(t, expected, readFile(t, p))
}

func TestUpsertBranchAnnotations_UpdatesExistingMatchingRepo(t *testing.T) {
	initial := "# header\norg/repo1\norg/repo2\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{
		"org/repo1": "feature-x",
	})
	assert.NoError(t, err)

	expected := "# header\norg/repo1 # branch=feature-x\norg/repo2\n"
	assert.Equal(t, expected, readFile(t, p))
}

func TestUpsertBranchAnnotations_PreservesFreeFormComment(t *testing.T) {
	initial := "org/repo1 # reviewer note\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{
		"org/repo1": "feature-x",
	})
	assert.NoError(t, err)

	// The existing comment text is preserved; the branch= token is injected.
	result := readFile(t, p)
	assert.Contains(t, result, "branch=feature-x")
	assert.Contains(t, result, "reviewer note")
}

func TestUpsertBranchAnnotations_InsertsSeparatorWhenCommentHasNone(t *testing.T) {
	// Parser accepts `org/repo#note` (no space after '#'). If we inject
	// `branch=X` without a separator we'd get `branch=Xnote` and the parser
	// would later read the branch as `Xnote`. Confirm we insert a space.
	initial := "org/repo1#reviewer note\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{"org/repo1": "feat"})
	assert.NoError(t, err)

	result := readFile(t, p)
	// The resulting comment must have whitespace between our branch= token
	// and the original note — the round-trip parse must yield branch="feat".
	assert.Contains(t, result, "branch=feat ")
	assert.Contains(t, result, "reviewer note")

	// Round-trip through the parser to confirm the branch is correctly read.
	branch := branchAnnotationRegexp.FindStringSubmatch(strings.SplitN(result, "#", 2)[1])
	if assert.NotNil(t, branch) {
		assert.Equal(t, "feat", branch[1])
	}
}

func TestUpsertBranchAnnotations_IdempotentOnMatchingBranch(t *testing.T) {
	initial := "org/repo1 # branch=feature-x\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{
		"org/repo1": "feature-x",
	})
	assert.NoError(t, err)
	assert.Equal(t, initial, readFile(t, p))
}

func TestUpsertBranchAnnotations_ErrorsOnConflict(t *testing.T) {
	initial := "org/repo1 # branch=feature-x\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{
		"org/repo1": "feature-y",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicting")
	// File must be untouched.
	assert.Equal(t, initial, readFile(t, p))
}

func TestUpsertBranchAnnotations_AtomicOnMixedConflicts(t *testing.T) {
	initial := "org/repo1 # branch=existing\norg/repo2\n"
	p := writeReposFile(t, initial)

	// repo1 conflicts; repo2 would be fine. We expect NO write to happen.
	err := UpsertBranchAnnotations(p, map[string]string{
		"org/repo1": "different",
		"org/repo2": "fine",
	})
	assert.Error(t, err)
	assert.Equal(t, initial, readFile(t, p))
}

func TestUpsertBranchAnnotations_PreservesFullLineCommentsAndBlankLines(t *testing.T) {
	initial := "# header comment\n\norg/repo1\n\n# mid comment\norg/repo2\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{
		"org/repo2": "fix",
	})
	assert.NoError(t, err)

	expected := "# header comment\n\norg/repo1\n\n# mid comment\norg/repo2 # branch=fix\n"
	assert.Equal(t, expected, readFile(t, p))
}

func TestUpsertBranchAnnotations_RoundTripIsIdentical(t *testing.T) {
	initial := "# header\norg/repo1 # branch=foo\n\norg/repo2\n"
	p := writeReposFile(t, initial)

	err := UpsertBranchAnnotations(p, map[string]string{}) // no changes
	assert.NoError(t, err)
	assert.Equal(t, initial, readFile(t, p))
}

func TestUpsertBranchAnnotations_MissingFileErrors(t *testing.T) {
	err := UpsertBranchAnnotations("/nonexistent/path/repos.txt", map[string]string{"org/r": "b"})
	assert.Error(t, err)
}
