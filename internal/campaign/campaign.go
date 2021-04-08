package campaign

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Repo struct {
	Host         string
	OrgName      string
	RepoName     string
	FullRepoName string
}

type CampaignDirectory struct {
	Repos []Repo
}

func OpenCampaignDirectory() (*CampaignDirectory, error) {
	file, err := os.Open("repos.txt")
	if err != nil {
		return nil, errors.New("Unable to open repos.txt file")
	}
	defer file.Close()

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
				return nil, errors.New(fmt.Sprintf("Unable to parse entry in repos.txt file: %s", line))
			}
			repos = append(repos, repo)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("Unable to open repos.txt file")
	}

	return &CampaignDirectory{
		Repos: repos,
	}, nil
}
