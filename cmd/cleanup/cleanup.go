package cleanup

import (
	"bufio"
	"github.com/skyscanner/turbolift/internal/campaign"
	"github.com/skyscanner/turbolift/internal/github"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var (
	gh          github.GitHub = github.NewRealGitHub()
	cleanupFile               = ".cleanup.txt"
	apply       bool
	repoFile    string
)

func NewCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleans up forks used in this campaign",
		Run:   run,
	}

	cmd.Flags().BoolVar(&apply, "apply", false, "Delete unused forks rather than just listing them")
	cmd.Flags().StringVar(&repoFile, "repos", "repos.txt", "A file containing a list of repositories to cleanup.")

	return cmd
}

func run(c *cobra.Command, _ []string) {
	if apply {
		logger := logging.NewLogger(c)
		if _, err := os.Stat(cleanupFile); os.IsNotExist(err) {
			logger.Errorf("The file %s does not exist. Please run `turbolift cleanup` without the --apply flag first.", cleanupFile)
		}
		readFileActivity := logger.StartActivity("Reading cleanup file")
		cleanupContents, err := os.Open(cleanupFile)
		if err != nil {
			readFileActivity.EndWithFailure(err)
			return
		}
		defer func(reposToDelete *os.File) {
			err := reposToDelete.Close()
			if err != nil {
				readFileActivity.EndWithFailure(err)
			}
		}(cleanupContents)
		scanner := bufio.NewScanner(cleanupContents)
		for scanner.Scan() {
			err = gh.DeleteFork(logger.Writer(), scanner.Text())
			if err != nil {
				readFileActivity.EndWithFailure(err)
				return
			}
		}
	} else {
		logger := logging.NewLogger(c)
		readCampaignActivity := logger.StartActivity("Reading campaign data (%s)", repoFile)
		options := campaign.NewCampaignOptions()
		options.RepoFilename = repoFile
		dir, err := campaign.OpenCampaign(options)
		if err != nil {
			readCampaignActivity.EndWithFailure(err)
			return
		}
		readCampaignActivity.EndWithSuccess()

		deletableForksActivity := logger.StartActivity("Checking for deletable forks")
		deletableForks, err := os.Create(cleanupFile)
		if err != nil {
			deletableForksActivity.EndWithFailure(err)
			return
		}
		defer func(deletableForks *os.File) {
			err := deletableForks.Close()
			if err != nil {
				deletableForksActivity.EndWithFailure(err)
			}
		}(deletableForks)
		var doneCount, errorCount int
		for _, repo := range dir.Repos {
			isFork, err := gh.IsFork(logger.Writer(), repo.FullRepoName)
			if err != nil {
				deletableForksActivity.EndWithFailure(err)
				errorCount++
				continue
			}
			repoDirPath := path.Join("work", repo.OrgName, repo.RepoName)
			pr, err := gh.GetPR(logger.Writer(), repoDirPath, dir.Name)
			if err != nil {
				deletableForksActivity.EndWithFailure(err)
				errorCount++
				continue
			}
			prClosed := pr.Closed == true
			if isFork && prClosed {
				deletableForks.WriteString(repo.FullRepoName + "\n")
			}
			doneCount++
		}
		logger.Printf("A list of forks used in this campaign has been written to %s. Check these carefully and run `turbolift cleanup --apply` in order to delete them.", cleanupFile)
		deletableForksActivity.EndWithSuccess()
	}
}
