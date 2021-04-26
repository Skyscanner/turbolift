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
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestItReadsSimpleRepoNamesFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "org/repo2")

	campaign, err := OpenCampaign()
	assert.NoError(t, err)

	assert.Equal(t, campaign.Name, testsupport.Pwd())
	assert.Equal(t, campaign.Repos, []Repo{
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
	})
	assert.Equal(t, campaign.PrTitle, "PR title")
	assert.Equal(t, campaign.PrBody, "PR body")
}

func TestItReadsRepoNamesWithOtherHostsFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "mygitserver.com/org/repo2")

	campaign, err := OpenCampaign()
	assert.NoError(t, err)

	assert.Equal(t, campaign.Name, testsupport.Pwd())
	assert.Equal(t, campaign.Repos, []Repo{
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
	})
	assert.Equal(t, campaign.PrTitle, "PR title")
	assert.Equal(t, campaign.PrBody, "PR body")
}

func TestItIgnoresCommentedLines(t *testing.T) {
	testsupport.PrepareTempCampaign(false, "org/repo1", "#org/repo2")

	campaign, err := OpenCampaign()
	assert.NoError(t, err)

	assert.Equal(t, campaign.Name, testsupport.Pwd())
	assert.Equal(t, campaign.Repos, []Repo{
		{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
	})
	assert.Equal(t, campaign.PrTitle, "PR title")
	assert.Equal(t, campaign.PrBody, "PR body")
}
