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

	dummyPrDescription := "# PR title\nPR body"
	err = ioutil.WriteFile("README.md", []byte(dummyPrDescription), os.ModePerm|0644)
	if err != nil {
		panic(err)
	}
}
