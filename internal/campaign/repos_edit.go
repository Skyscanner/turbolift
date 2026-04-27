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
	"fmt"
	"os"
	"sort"
	"strings"
)

// UpsertBranchAnnotations rewrites repos.txt so that each repo in `branches`
// ends up annotated with its PR branch. The write is atomic in the sense that
// if any repo already has a conflicting `branch=<other>` annotation, the
// function returns an error WITHOUT modifying the file. Callers are expected
// to surface conflicts to the user for manual resolution.
//
// Preserved verbatim: full-line comments, blank lines, the order of existing
// lines, and any non-`branch=` text in trailing comments on matched lines.
// Repos in `branches` that aren't already in the file are appended at the
// end, in alphabetical order for deterministic output.
func UpsertBranchAnnotations(path string, branches map[string]string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read %s: %w", path, err)
	}

	// Preserve original trailing newline behaviour by splitting without
	// stripping and handling the final empty segment specially.
	lines := strings.Split(string(raw), "\n")
	hasTrailingNewline := len(lines) > 0 && lines[len(lines)-1] == ""
	if hasTrailingNewline {
		lines = lines[:len(lines)-1]
	}

	// First pass: determine which repos we need to update and collect
	// conflicts. No writes yet — we want all-or-nothing atomicity so a
	// conflict doesn't leave repos.txt half-updated.
	type update struct {
		lineIdx    int
		newContent string
	}
	var updates []update
	remaining := make(map[string]string, len(branches))
	for k, v := range branches {
		remaining[k] = v
	}
	var conflicts []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Extract the repo part (before any '#'). Same split rule as the
		// parser: repo names cannot contain '#'.
		var repoPart, commentPart string
		var hadComment bool
		if idx := strings.Index(line, "#"); idx >= 0 {
			repoPart = line[:idx]
			commentPart = line[idx+1:]
			hadComment = true
		} else {
			repoPart = line
		}
		repoKey := strings.TrimSpace(repoPart)

		wantBranch, want := branches[repoKey]
		if !want {
			continue
		}
		delete(remaining, repoKey)

		// Detect existing `branch=<x>` in the comment.
		existingMatch := branchAnnotationRegexp.FindStringSubmatch(commentPart)
		if existingMatch != nil {
			existingBranch := existingMatch[1]
			if existingBranch == wantBranch {
				// Already correct — leave line untouched.
				continue
			}
			conflicts = append(conflicts, fmt.Sprintf("%s: existing branch=%s, want branch=%s", repoKey, existingBranch, wantBranch))
			continue
		}

		// Inject `branch=<wantBranch>` into the line. If there's an existing
		// comment without a branch annotation, prepend our annotation and
		// keep the rest of the comment text.
		var newLine string
		if !hadComment {
			newLine = fmt.Sprintf("%s # branch=%s", repoKey, wantBranch)
		} else {
			newLine = fmt.Sprintf("%s # branch=%s%s", repoKey, wantBranch, commentPart)
		}
		updates = append(updates, update{lineIdx: i, newContent: newLine})
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("conflicting branch annotations in %s (resolve manually and retry):\n  %s", path, strings.Join(conflicts, "\n  "))
	}

	// Apply in-memory updates and append new lines for unmatched repos.
	for _, u := range updates {
		lines[u.lineIdx] = u.newContent
	}
	if len(remaining) > 0 {
		keys := make([]string, 0, len(remaining))
		for k := range remaining {
			keys = append(keys, k)
		}
		// Sort so behaviour doesn't depend on map iteration order and
		// subsequent runs produce deterministic diffs.
		sort.Strings(keys)
		for _, k := range keys {
			lines = append(lines, fmt.Sprintf("%s # branch=%s", k, remaining[k]))
		}
	}

	out := strings.Join(lines, "\n")
	if hasTrailingNewline || len(remaining) > 0 {
		out += "\n"
	}

	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", path, err)
	}
	return nil
}
