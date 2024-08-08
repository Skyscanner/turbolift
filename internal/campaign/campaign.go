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
 *
 */

package campaign

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	//go:embed templates/.gitignore
	gitignoreTemplate string

	//go:embed templates/.turbolift
	turboliftTemplate string

	//go:embed templates/README.md
	readmeTemplate string

	//go:embed templates/repos.txt
	reposTemplate string
)

type TemplateVariables struct {
	CampaignName string
}

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

func OpenCampaign(options *CampaignOptions) (*Campaign, error) {
	dir, _ := os.Getwd()
	dirBasename := filepath.Base(dir)

	repos, err := readReposTxtFile(options.RepoFilename)
	if err != nil {
		return nil, err
	}

	prTitle, prBody, err := readPrDescriptionFile(options.PrDescriptionFilename)
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
	uniq := map[string]interface{}{}
	var repos []Repo
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && len(line) > 0 {
			if _, seen := uniq[line]; seen {
				continue
			}
			uniq[line] = struct{}{}

			splitLine := strings.Split(line, "/")
			numParts := len(splitLine)

			var repo Repo
			switch numParts {
			case 2:
				repo = Repo{
					OrgName:      splitLine[0],
					RepoName:     splitLine[1],
					FullRepoName: line,
				}
			case 3:
				repo = Repo{
					Host:         splitLine[0],
					OrgName:      splitLine[1],
					RepoName:     splitLine[2],
					FullRepoName: line,
				}
			default:
				return nil, fmt.Errorf("unable to parse entry in %s file: %s", filename, line)
			}
			repos = append(repos, repo)
		}
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

func CreateInitialFiles(campaignName string) error {
	data := TemplateVariables{
		CampaignName: campaignName,
	}
	files := map[string]string{
		".gitignore": gitignoreTemplate,
		".turbolift": turboliftTemplate,
		"README.md":  readmeTemplate,
		"repos.txt":  reposTemplate,
	}
	for filename, templateFile := range files {
		err := applyTemplate(filepath.Join(campaignName, filename), templateFile, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// Applies a given template and data to produce a file with the outputFilename
func applyTemplate(outputFilename string, templateContent string, data interface{}) error {
	file, err := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("unable to open file for output: %w", err)
	}

	defer file.Close()

	parsedTemplate, err := template.New("").Parse(templateContent)

	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	err = parsedTemplate.Execute(file, data)

	if err != nil {
		return fmt.Errorf("unable to write templated file: %w", err)
	}
	return nil
}

func ApplyReadMeTemplate(filename string, dirName string) error {
	data := TemplateVariables{
		CampaignName: dirName,
	}
	err := applyTemplate(filename, readmeTemplate, data)
	if err != nil {
		return err
	}
	return nil
}

func PrDescriptionUnchanged(dir *Campaign) (bool, error) {
	tempPrDescriptionFile, err := os.Create("turbolift-temp-pr-description-*")
	if err != nil {
		return false, fmt.Errorf("unable to create temp pr description file: %w", err)
	}
	err = ApplyReadMeTemplate(tempPrDescriptionFile.Name(), dir.Name)
	if err != nil {
		return false, fmt.Errorf("unable to write temp pr description file: %w", err)
	}

	defer os.Remove(tempPrDescriptionFile.Name())

	originalPrTitle, originalPrBody, err := readPrDescriptionFile(tempPrDescriptionFile.Name())
	if err != nil {
		return false, fmt.Errorf("unable to read pr description file: %w", err)
	}
	return dir.PrTitle == originalPrTitle || dir.PrBody == originalPrBody || dir.PrTitle == "", err
}
