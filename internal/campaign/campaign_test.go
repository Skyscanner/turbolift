package campaign

import (
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestItReadsSimpleRepoNamesFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaignDirectory("org/repo1", "org/repo2")

	campaignDirectory, err := OpenCampaignDirectory()
	assert.NoError(t, err)

	assert.Equal(t, campaignDirectory.Name, testsupport.Pwd())
	assert.Equal(t, campaignDirectory.Repos, []Repo{
		Repo{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
		Repo{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo2",
			FullRepoName: "org/repo2",
		},
	})
}

func TestItReadsRepoNamesWithOtherHostsFromReposFile(t *testing.T) {
	testsupport.PrepareTempCampaignDirectory("org/repo1", "mygitserver.com/org/repo2")

	campaignDirectory, err := OpenCampaignDirectory()
	assert.NoError(t, err)

	assert.Equal(t, campaignDirectory.Name, testsupport.Pwd())
	assert.Equal(t, campaignDirectory.Repos, []Repo{
		Repo{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
		Repo{
			Host:         "mygitserver.com",
			OrgName:      "org",
			RepoName:     "repo2",
			FullRepoName: "mygitserver.com/org/repo2",
		},
	})
}

func TestItIgnoresCommentedLines(t *testing.T) {
	testsupport.PrepareTempCampaignDirectory("org/repo1", "#org/repo2")

	campaignDirectory, err := OpenCampaignDirectory()
	assert.NoError(t, err)

	assert.Equal(t, campaignDirectory.Name, testsupport.Pwd())
	assert.Equal(t, campaignDirectory.Repos, []Repo{
		Repo{
			Host:         "",
			OrgName:      "org",
			RepoName:     "repo1",
			FullRepoName: "org/repo1",
		},
	})
}
