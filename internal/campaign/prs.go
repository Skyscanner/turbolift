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
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// PRRef identifies a specific pull request by (optional host, org, repo, number).
// Host is empty for PRs on github.com; populated for GitHub Enterprise hosts
// discovered from URL-form input.
type PRRef struct {
	Host     string
	OrgName  string
	RepoName string
	Number   int
}

// FullRepoName returns the org/repo identifier (without host prefix), suitable
// for use as a key when looking up or storing repos in repos.txt. The host is
// deliberately omitted because repos.txt lines may or may not carry the host
// prefix — downstream callers that need it can read PRRef.Host explicitly.
func (p PRRef) FullRepoName() string {
	return p.OrgName + "/" + p.RepoName
}

// shorthandRegexp matches `org/repo#N`. We accept alphanumerics, dash, dot,
// and underscore in org/repo names — the set of characters GitHub permits.
var shorthandRegexp = regexp.MustCompile(`^([A-Za-z0-9._-]+)/([A-Za-z0-9._-]+)#(\d+)$`)

// ParsePRRef accepts either a full URL (`https://host/org/repo/pull/N`) or the
// shorthand `org/repo#N`. Returns an error for anything else — the caller is
// responsible for surfacing malformed input to the user loudly, since silent
// skipping would be surprising.
func ParsePRRef(line string) (PRRef, error) {
	line = strings.TrimSpace(line)

	// URL form: parse with net/url so we correctly handle alternate hosts
	// (GHE) and any path quirks.
	if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		u, err := url.Parse(line)
		if err != nil {
			return PRRef{}, fmt.Errorf("invalid PR URL %q: %w", line, err)
		}
		// Expect path like /org/repo/pull/N.
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) != 4 || parts[2] != "pull" {
			return PRRef{}, fmt.Errorf("PR URL does not look like /org/repo/pull/N: %q", line)
		}
		num, err := strconv.Atoi(parts[3])
		if err != nil || num <= 0 {
			return PRRef{}, fmt.Errorf("PR URL has no numeric PR number: %q", line)
		}
		host := u.Host
		// github.com is the default — leave Host empty so downstream logic
		// can treat "no host" and "github.com" identically.
		if host == "github.com" {
			host = ""
		}
		return PRRef{Host: host, OrgName: parts[0], RepoName: parts[1], Number: num}, nil
	}

	// Shorthand form.
	if m := shorthandRegexp.FindStringSubmatch(line); m != nil {
		num, _ := strconv.Atoi(m[3]) // regex guarantees digits
		return PRRef{OrgName: m[1], RepoName: m[2], Number: num}, nil
	}

	return PRRef{}, fmt.Errorf("unrecognised PR reference %q (expected URL or org/repo#N)", line)
}

// ReadPRsFile reads a file containing one PR reference per line, skipping
// blank lines and full-line `#` comments. Exact duplicates (same org/repo/N)
// are silently deduped. Multiple PRs against the same repo are an error —
// turbolift assimilates at most one PR per repo per campaign.
func ReadPRsFile(path string) ([]PRRef, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open PRs file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	seenKey := map[string]PRRef{} // key = "org/repo"
	var prs []PRRef
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pr, err := ParsePRRef(line)
		if err != nil {
			return nil, err
		}
		key := pr.FullRepoName()
		if existing, seen := seenKey[key]; seen {
			if existing.Number == pr.Number {
				continue // exact dup — silent skip
			}
			return nil, fmt.Errorf("multiple PRs for %s in %s (#%d and #%d)", key, path, existing.Number, pr.Number)
		}
		seenKey[key] = pr
		prs = append(prs, pr)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to read PRs file %s: %w", path, err)
	}
	return prs, nil
}
