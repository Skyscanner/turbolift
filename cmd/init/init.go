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
	_ "embed"
	"fmt"
	"github.com/skyscanner/turbolift/internal/colors"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
	"html/template"
	"os"
	"path/filepath"
)

var (
	campaignName string

	//go:embed templates/.gitignore
	gitignoreTemplate string

	//go:embed templates/.turbolift
	turboliftTemplate string

	//go:embed templates/README.md
	readmeTemplate string

	//go:embed templates/repos.txt
	reposTemplate string
)

type TemplateVariables struct {
	CampaignName string
}

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Turbolift campaign directory",
		Run:   run,
	}

	cmd.Flags().StringVarP(&campaignName, "name", "n", "", "Campaign name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func run(c *cobra.Command, _ []string) {
	logger := logging.NewLogger(c)

	createDirActivity := logger.StartActivity("Creating work directory")
	// Create a directory for both the campaign and its nested work directory
	workDirectory := filepath.Join(campaignName, "work")
	err := os.MkdirAll(workDirectory, os.ModeDir|0755)

	if err != nil {
		createDirActivity.EndWithFailure(err)
		return
	} else {
		createDirActivity.EndWithSuccess()
	}

	createFilesActivity := logger.StartActivity("Creating initial files")
	data := TemplateVariables{
		CampaignName: campaignName,
	}

	files := map[string]string{
		".gitignore": gitignoreTemplate,
		".turbolift": turboliftTemplate,
		"README.md":  readmeTemplate,
		"repos.txt":  reposTemplate,
	}
	for filename, templateFile := range files {
		err := applyTemplate(filepath.Join(campaignName, filename), templateFile, data)
		if err != nil {
			createFilesActivity.EndWithFailure(err)
			return
		}
	}
	createFilesActivity.EndWithSuccess()

	logger.Successf("turbolift init is done - next:\n")
	logger.Println("1. Run", colors.Cyan("cd ", campaignName))
	logger.Println("2. Update repos.txt with the names of the repos that need changing (either manually or using a tool to generate a list of repos)")
	logger.Println("3. Run", colors.Cyan("turbolift clone"))
}

// Applies a given template and data to produce a file with the outputFilename
func applyTemplate(outputFilename string, templateContent string, data interface{}) error {
	readme, err := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("Unable to open file for output: %w", err)
	}

	parsedTemplate, err := template.New("").Parse(templateContent)

	if err != nil {
		return fmt.Errorf("Unable to parse template: %w", err)
	}

	err = parsedTemplate.Execute(readme, data)

	if err != nil {
		return fmt.Errorf("Unable to write templated file: %w", err)
	}
	return nil
}
