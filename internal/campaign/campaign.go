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
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const CampaignPrefix = "turbolift-"

type Repo struct {
	Host         string
	OrgName      string
	RepoName     string
	FullRepoName string
	// Branch is the git branch to operate on for this repo. Defaults to the
	// campaign name during OpenCampaign when no `# branch=...` annotation is
	// present in repos.txt. Populated explicitly when `clone --from-prs`
	// assimilates existing PRs with non-matching head refs.
	Branch string
}

// branchAnnotationRegexp matches a whitespace-bounded `branch=<value>` token
// within a trailing comment in repos.txt. Any other text in the comment is
// free-form and ignored by the tool, so users can combine branch annotations
// with arbitrary notes.
var branchAnnotationRegexp = regexp.MustCompile(`\bbranch=(\S+)`)

type Campaign struct {
	Name    string
	Repos   []Repo
	PrTitle string
	PrBody  string
}

func (r Repo) FullRepoPath() string {
	return path.Join("work", r.OrgName, r.RepoName) // i.e. work/org/repo
}

type CampaignOptions struct {
	RepoFilename          string
	PrDescriptionFilename string
}

func NewCampaignOptions() *CampaignOptions {
	return &CampaignOptions{
		RepoFilename:          "repos.txt",
		PrDescriptionFilename: "README.md",
	}
}

func ApplyCampaignNamePrefix(name string) string {
	if strings.HasPrefix(name, CampaignPrefix) {
		return name
	}
	return fmt.Sprintf("%s%s", CampaignPrefix, name)
}

func OpenCampaign(options *CampaignOptions) (*Campaign, error) {
	dir, _ := os.Getwd()
	dirBasename := filepath.Base(dir)
	name := ApplyCampaignNamePrefix(dirBasename)

	repos, err := readReposTxtFile(options.RepoFilename)
	if err != nil {
		return nil, err
	}

	prTitle, prBody, err := readPrDescriptionFile(options.PrDescriptionFilename)
	if err != nil {
		return nil, err
	}

	// Ensure every repo has a non-empty Branch. Downstream commands can then
	// use repo.Branch unconditionally without having to fall back to the
	// campaign name — and `--from-prs` callers that populate Branch from PR
	// head refs take precedence over this default.
	for i := range repos {
		if repos[i].Branch == "" {
			repos[i].Branch = name
		}
	}

	return &Campaign{
		Name:    name,
		Repos:   repos,
		PrTitle: prTitle,
		PrBody:  prBody,
	}, nil
}

func readReposTxtFile(filename string) ([]Repo, error) {
	if filename == "" {
		return nil, errors.New("no repos filename to open")
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open repo file: %s", filename)
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	// branchByRepo tracks the branch annotation we've seen for each repo.
	// We dedupe identical-in-every-way entries silently, but we error loudly
	// when the same repo appears with conflicting branch annotations — that
	// almost certainly indicates a bug in whoever wrote the file.
	branchByRepo := map[string]string{}
	var repos []Repo
	for scanner.Scan() {
		line := scanner.Text()
		// Full-line comments and blank lines are ignored as before.
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Split on the first '#' to separate the repo name from any trailing
		// comment. Repo names cannot contain '#', so this split is unambiguous.
		var repoPart, commentPart string
		if idx := strings.Index(line, "#"); idx >= 0 {
			repoPart = line[:idx]
			commentPart = line[idx+1:]
		} else {
			repoPart = line
		}
		repoPart = strings.TrimSpace(repoPart)
		if repoPart == "" {
			continue
		}

		splitLine := strings.Split(repoPart, "/")
		numParts := len(splitLine)
		var repo Repo
		switch numParts {
		case 2:
			repo = Repo{
				OrgName:      splitLine[0],
				RepoName:     splitLine[1],
				FullRepoName: repoPart,
			}
		case 3:
			repo = Repo{
				Host:         splitLine[0],
				OrgName:      splitLine[1],
				RepoName:     splitLine[2],
				FullRepoName: repoPart,
			}
		default:
			return nil, fmt.Errorf("unable to parse entry in %s file: %s", filename, line)
		}

		// Scan the trailing comment for a `branch=<value>` token. Other
		// comment text is free-form and ignored so users can keep notes.
		if m := branchAnnotationRegexp.FindStringSubmatch(commentPart); m != nil {
			repo.Branch = m[1]
		}

		if existing, seen := branchByRepo[repo.FullRepoName]; seen {
			if existing != repo.Branch {
				return nil, fmt.Errorf("conflicting branch annotations for %s: %q vs %q", repo.FullRepoName, existing, repo.Branch)
			}
			continue
		}
		branchByRepo[repo.FullRepoName] = repo.Branch
		repos = append(repos, repo)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to open %s file: %w", filename, err)
	}

	return repos, nil
}

func readPrDescriptionFile(filename string) (string, string, error) {
	if filename == "" {
		return "", "", errors.New("no PR description file to open")
	}
	file, err := os.Open(filename)
	if err != nil {
		return "", "", fmt.Errorf("unable to open PR description file: %s", filename)
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	prTitle := ""
	prBodyLines := []string{}
	for scanner.Scan() {
		line := scanner.Text()

		if prTitle == "" {
			trimmedFirstLine := strings.TrimLeft(line, "# ")
			prTitle = trimmedFirstLine
		} else {
			prBodyLines = append(prBodyLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("unable to read PR description file: %s", filename)
	}

	return prTitle, strings.Join(prBodyLines, "\n"), nil
}
