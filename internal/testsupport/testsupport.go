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

func PrepareTempCampaign(createDirs bool, repos ...string) {
	CreateAndEnterTempDirectory()

	delimitedList := strings.Join(repos, "\n")
	err := ioutil.WriteFile("repos.txt", []byte(delimitedList), os.ModePerm|0644)
	if err != nil {
		panic(err)
	}

	if createDirs {
		for _, name := range repos {
			dirToCreate := path.Join("work", name)
			err := os.MkdirAll(dirToCreate, os.ModeDir|0755)
			if err != nil {
				panic(err)
			}
		}
	}

	dummyPrDescription := `# PR title
	PR body`
	err = ioutil.WriteFile("README.md", []byte(dummyPrDescription), os.ModePerm|0644)
	if err != nil {
		panic(err)
	}
}
