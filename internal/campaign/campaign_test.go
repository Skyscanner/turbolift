package campaign

import (
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestItReadsSimpleRepoNamesFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaignDirectory(false, "org/repo1", "org/repo2")

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
}

func TestItReadsRepoNamesWithOtherHostsFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaignDirectory(false, "org/repo1", "mygitserver.com/org/repo2")

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
}

func TestItIgnoresCommentedLines(t *testing.T) {
	testsupport.PrepareTempCampaignDirectory(false, "org/repo1", "#org/repo2")

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
}
