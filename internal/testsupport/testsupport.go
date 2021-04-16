package testsupport

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func Pwd() string {
	dir, _ := os.Getwd()
	return path.Base(dir)
}

func CreateAndEnterTempDirectory() {
	tempDir, _ := ioutil.TempDir("", "turbolift-test-*")
	err := os.Chdir(tempDir)

	if err != nil {
		panic(err)
	}
}

func PrepareTempCampaignDirectory(repos ...string) {
	CreateAndEnterTempDirectory()

	delimitedList := strings.Join(repos, "\n")
	err := ioutil.WriteFile("repos.txt", []byte(delimitedList), os.ModePerm|0644)
	if err != nil {
		panic(err)
	}
}
