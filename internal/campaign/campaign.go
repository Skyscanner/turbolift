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

package campaign

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

type Repo struct {
	Host         string
	OrgName      string
	RepoName     string
	FullRepoName string
}

type Campaign struct {
	Name    string
	Repos   []Repo
	PrTitle string
	PrBody  string
}

func OpenCampaign() (*Campaign, error) {
	dir, _ := os.Getwd()
	dirBasename := path.Base(dir)

	repos, err := readReposTxtFile()
	if err != nil {
		return nil, err
	}

	prTitle, prBody, err := readPrDescriptionFile()
	if err != nil {
		return nil, err
	}

	return &Campaign{
		Name:    dirBasename,
		Repos:   repos,
		PrTitle: prTitle,
		PrBody:  prBody,
	}, nil
}

func readReposTxtFile() ([]Repo, error) {
	file, err := os.Open("repos.txt")
	if err != nil {
		return nil, errors.New("Unable to open repos.txt file")
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	var repos []Repo
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			splitLine := strings.Split(line, "/")
			numParts := len(splitLine)

			var repo Repo
			if numParts == 2 {
				repo = Repo{
					OrgName:      splitLine[0],
					RepoName:     splitLine[1],
					FullRepoName: line,
				}
			} else if numParts == 3 {
				repo = Repo{
					Host:         splitLine[0],
					OrgName:      splitLine[1],
					RepoName:     splitLine[2],
					FullRepoName: line,
				}
			} else {
				return nil, fmt.Errorf("Unable to parse entry in repos.txt file: %s", line)
			}
			repos = append(repos, repo)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Unable to open repos.txt file: %w", err)
	}

	return repos, nil
}

func readPrDescriptionFile() (string, string, error) {
	file, err := os.Open("README.md")
	if err != nil {
		return "", "", errors.New("Unable to open README.md file")
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
		return "", "", errors.New("Unable to read README.md file")
	}

	return prTitle, strings.Join(prBodyLines, "\n"), nil
}
