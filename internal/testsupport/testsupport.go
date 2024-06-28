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

package testsupport

import (
	"fmt"
	"github.com/skyscanner/turbolift/internal/campaign"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Pwd() string {
	dir, _ := os.Getwd()
	return filepath.Base(dir)
}

func CreateAndEnterTempDirectory() string {
	tempDir, _ := ioutil.TempDir("", "turbolift-test-*")
	err := os.Chdir(tempDir)
	if err != nil {
		panic(err)
	}
	return tempDir
}

func PrepareTempCampaign(createDirs bool, repos ...string) string {
	tempDir := CreateAndEnterTempDirectory()

	delimitedList := strings.Join(repos, "\n")
	err := ioutil.WriteFile("repos.txt", []byte(delimitedList), os.ModePerm|0o644)
	if err != nil {
		panic(err)
	}

	if createDirs {
		for _, name := range repos {
			dirToCreate := path.Join("work", name)
			err := os.MkdirAll(dirToCreate, os.ModeDir|0o755)
			if err != nil {
				panic(err)
			}
		}
	}

	dummyPrDescription := "# PR title\nPR body"
	err = ioutil.WriteFile("README.md", []byte(dummyPrDescription), os.ModePerm|0o644)
	if err != nil {
		panic(err)
	}

	return tempDir
}

func CreateAnotherRepoFile(filename string, repos ...string) {
	delimitedList := strings.Join(repos, "\n")
	err := ioutil.WriteFile(filename, []byte(delimitedList), os.ModePerm|0o644)
	if err != nil {
		panic(err)
	}
}

func CreateOrUpdatePrDescriptionFile(filename string, prTitle string, prBody string) {
	prDescription := fmt.Sprintf("# %s\n%s", prTitle, prBody)
	err := os.WriteFile(filename, []byte(prDescription), os.ModePerm|0o644)
	if err != nil {
		panic(err)
	}
}

func UseDefaultPrDescription(dirName string) {
	fileName := "README.md"
	err := os.Remove(fileName)
	if err != nil {
		panic(err)
	}
	err = campaign.ApplyReadMeTemplate(fileName, dirName)
	if err != nil {
		panic(err)
	}
}

func UseDefaultPrTitleOnly(dirName string) {
	UseDefaultPrDescription(dirName)
	// append some text to change the pr description body
	f, err := os.OpenFile("README.md", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	_, err = f.WriteString("additional pr description")
	if err != nil {
		panic(err)
	}
}

func UseDefaultPrBodyOnly(dirName string) {
	UseDefaultPrDescription(dirName)
	//	append something to first line to change title
	fileName := "README.md"
	content, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	lines[0] += "updated title"

	newContent := strings.Join(lines, "\n")

	err = os.WriteFile(fileName, []byte(newContent), 0644)
	if err != nil {
		panic(err)
	}
}
