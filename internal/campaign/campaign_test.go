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
	"testing"

	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestItReadsSimpleRepoNamesFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	defaultOptions := NewCampaignOptions()
	campaign, err := OpenCampaign(defaultOptions)
	assert.NoError(t, err)

	assert.Equal(t, testsupport.Pwd(), campaign.Name)
	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo2",
			FullRepoName: "org/repo2",
		},
	}, campaign.Repos)
	assert.Equal(t, "PR title", campaign.PrTitle)
	assert.Equal(t, "PR body", campaign.PrBody)
}

func TestItReadsRepoNamesWithOtherHostsFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "mygitserver.com/org/repo2")

	defaultOptions := NewCampaignOptions()
	campaign, err := OpenCampaign(defaultOptions)
	assert.NoError(t, err)

	assert.Equal(t, testsupport.Pwd(), campaign.Name)
	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
		{
			Host:         "mygitserver.com",
			OrgName:      "org",
			RepoName:     "repo2",
			FullRepoName: "mygitserver.com/org/repo2",
		},
	}, campaign.Repos)
	assert.Equal(t, "PR title", campaign.PrTitle)
	assert.Equal(t, "PR body", campaign.PrBody)
}

func TestItIgnoresCommentedLines(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "#org/repo2")

	defaultOptions := NewCampaignOptions()
	campaign, err := OpenCampaign(defaultOptions)
	assert.NoError(t, err)

	assert.Equal(t, testsupport.Pwd(), campaign.Name)
	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
	}, campaign.Repos)
	assert.Equal(t, "PR title", campaign.PrTitle)
	assert.Equal(t, "PR body", campaign.PrBody)
}

func TestItIgnoresEmptyLines(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "")

	defaultOptions := NewCampaignOptions()
	campaign, err := OpenCampaign(defaultOptions)
	assert.NoError(t, err)

	assert.Equal(t, testsupport.Pwd(), campaign.Name)
	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
	}, campaign.Repos)
	assert.Equal(t, "PR title", campaign.PrTitle)
	assert.Equal(t, "PR body", campaign.PrBody)
}

func TestItIgnoresEmptyAndCommentedLines(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "#Comment", "org/repo1", "")

	defaultOptions := NewCampaignOptions()
	campaign, err := OpenCampaign(defaultOptions)
	assert.NoError(t, err)

	assert.Equal(t, testsupport.Pwd(), campaign.Name)
	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
	}, campaign.Repos)
	assert.Equal(t, "PR title", campaign.PrTitle)
	assert.Equal(t, "PR body", campaign.PrBody)
}

func TestItIgnoresDuplicatedLines(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo1")

	defaultOptions := NewCampaignOptions()
	campaign, err := OpenCampaign(defaultOptions)
	assert.NoError(t, err)

	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
	}, campaign.Repos)
}

func TestItIgnoresDuplicatedNonSequentialLines(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2", "org/repo1")

	defaultOptions := NewCampaignOptions()
	campaign, err := OpenCampaign(defaultOptions)
	assert.NoError(t, err)

	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo2",
			FullRepoName: "org/repo2",
		},
	}, campaign.Repos)
}

func TestItShouldAcceptADifferentRepoFileSuccess(t *testing.T) {
	testsupport.PrepareTempCampaign(false)

	testsupport.CreateAnotherRepoFile("newrepos.txt", "org/repo1", "org/repo2", "org/repo3")
	options := NewCampaignOptions()
	options.RepoFilename = "newrepos.txt"
	campaign, err := OpenCampaign(options)
	assert.NoError(t, err)

	assert.Equal(t, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo2",
			FullRepoName: "org/repo2",
		},
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo3",
			FullRepoName: "org/repo3",
		},
	}, campaign.Repos)
}

func TestItShouldAcceptADifferentRepoFileNotExist(t *testing.T) {
	testsupport.PrepareTempCampaign(false)

	options := NewCampaignOptions()
	options.RepoFilename = "newrepos.txt"
	_, err := OpenCampaign(options)
	assert.Error(t, err)
}

func TestItShouldErrorWhenRepoFileIsEmpty(t *testing.T) {
	testsupport.PrepareTempCampaign(false)

	options := &CampaignOptions{}
	_, err := OpenCampaign(options)
	assert.Error(t, err)
}
