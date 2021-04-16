package commit

import (
	"bytes"
	"errors"
	"github.com/skyscanner/turbolift/internal/git"
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestItCommitsAllWithChanges(t *testing.T) {
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory(true, "org/repo1", "org/repo2")

	out, err := runCommand("some test message", []string{}...)
	assert.NoError(t, err)
	assert.Contains(t, out, "2 OK")

	fakeGit.AssertCalledWith(t, [][]string{
		{"isRepoChanged", "work/org/repo1"},
		{"commit", "work/org/repo1", "some test message"},
		{"isRepoChanged", "work/org/repo2"},
		{"commit", "work/org/repo2", "some test message"},
	})
}

func TestItSkipsReposWithoutChanges(t *testing.T) {
	fakeGit := git.NewFakeGit(func(output io.Writer, call []string) (bool, error) {
		if call[0] == "isRepoChanged" && call[1] == "work/org/repo1" {
			return false, nil
		} else {
			return true, nil
		}
	})
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory(true, "org/repo1", "org/repo2")

	out, err := runCommand("some test message", []string{}...)
	assert.NoError(t, err)
	assert.Contains(t, out, "No changes in org/repo1 - skipping commit")
	assert.Contains(t, out, "1 OK, 1 skipped")

	fakeGit.AssertCalledWith(t, [][]string{
		{"isRepoChanged", "work/org/repo1"},
		{"isRepoChanged", "work/org/repo2"},
		{"commit", "work/org/repo2", "some test message"},
	})
}

func TestItSkipsReposWhichErrorOnStatusChekc(t *testing.T) {
	fakeGit := git.NewFakeGit(func(output io.Writer, call []string) (bool, error) {
		if call[0] == "isRepoChanged" && call[1] == "work/org/repo1" {
			return false, errors.New("synthetic error")
		} else {
			return true, nil
		}
	})
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory(true, "org/repo1", "org/repo2")

	out, err := runCommand("some test message", []string{}...)
	assert.NoError(t, err)
	assert.Contains(t, out, "Error when checking for changes in org/repo1")
	assert.Contains(t, out, "1 OK, 0 skipped, 1 errored")

	fakeGit.AssertCalledWith(t, [][]string{
		{"isRepoChanged", "work/org/repo1"},
		{"isRepoChanged", "work/org/repo2"},
		{"commit", "work/org/repo2", "some test message"},
	})
}

func TestItSkipsMissingRepos(t *testing.T) {
	fakeGit := git.NewAlwaysSucceedsFakeGit()
	g = fakeGit

	testsupport.PrepareTempCampaignDirectory(true, "org/repo1", "org/repo2")
	err := os.RemoveAll("work/org/repo1")
	if err != nil {
		panic(err)
	}

	out, err := runCommand("some test message", []string{}...)
	assert.NoError(t, err)
	assert.Contains(t, out, "1 OK, 1 skipped")

	fakeGit.AssertCalledWith(t, [][]string{
		{"isRepoChanged", "work/org/repo2"},
		{"commit", "work/org/repo2", "some test message"},
	})
}

func runCommand(m string, args ...string) (string, error) {
	cmd := NewCommitCmd()
	outBuffer := bytes.NewBufferString("")
	cmd.SetOut(outBuffer)
	cmd.SetArgs(args)
	err := cmd.Flags().Set("message", m)
	if err != nil {
		panic(err)
	}

	err = cmd.Execute()

	if err != nil {
		return outBuffer.String(), err
	}
	return outBuffer.String(), nil
}
