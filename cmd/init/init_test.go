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

package init

import (
	"github.com/skyscanner/turbolift/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestAllFilesAreCreated(t *testing.T) {
	testsupport.CreateAndEnterTempDirectory()
	runCommand()

	assert.DirExistsf(t, "foo", "campaign directory should have been created")
	assert.DirExistsf(t, "foo/work", "work directory should have been created")
	assert.FileExists(t, "foo/.gitignore", "a .gitignore file should have been created")
	assert.FileExists(t, "foo/.turbolift", "a .turbolift file should have been created")
	assert.FileExists(t, "foo/README.md", "a README.md file should have been created")
	assert.FileExists(t, "foo/repos.txt", "a repos.txt file should have been created")
}

func TestTemplatedFilesHaveExpectedContent(t *testing.T) {
	testsupport.CreateAndEnterTempDirectory()
	runCommand()

	readmeContents, err := ioutil.ReadFile("foo/README.md")
	if err != nil {
		panic(err)
	}

	// Don't be too specific about expected content, to avoid test fragility
	assert.Contains(t, string(readmeContents), "foo")
}

func runCommand() {
	cmd := NewInitCmd()
	cmd.SetArgs([]string{"--name", "foo"})
	err := cmd.Execute()

	if err != nil {
		panic(err)
	}
}
