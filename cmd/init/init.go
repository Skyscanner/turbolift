package init

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/skyscanner/turbolift/internal/colors"
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

func CreateInitCmd() *cobra.Command {
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
	// Create a directory for both the campaign and its nested work directory
	workDirectory := filepath.Join(campaignName, "work")
	err := os.MkdirAll(workDirectory, os.ModeDir|0755)

	if err != nil {
		c.Println("Unable to create directory ", workDirectory, ": ", err)
	}

	data := TemplateVariables{
		CampaignName: campaignName,
	}

	applyTemplate(filepath.Join(campaignName, ".gitignore"), gitignoreTemplate, data)
	applyTemplate(filepath.Join(campaignName, ".turbolift"), turboliftTemplate, data)
	applyTemplate(filepath.Join(campaignName, "README.md"), readmeTemplate, data)
	applyTemplate(filepath.Join(campaignName, "repos.txt"), reposTemplate, data)

	c.Println(colors.Green("✅ turbolift init is done - next:"))
	c.Println("1. Run", colors.Cyan("cd ", campaignName))
	c.Println("2. Update repos.txt with the names of the repos that need changing (either manually or using a tool to generate a list of repos)")
	c.Println("3. Run", colors.Cyan("turbolift clone"))
}

// Applies a given template and data to produce a file with the outputFilename
func applyTemplate(outputFilename string, templateContent string, data interface{}) error {
	readme, err := os.Create(outputFilename)

	parsedTemplate, err := template.New("").Parse(templateContent)

	if err != nil {
		return errors.New("Unable to parse template")
	}

	err = parsedTemplate.Execute(readme, data)

	if err != nil {
		return errors.New(fmt.Sprintf("Unable to write templated file: %s", err))
	}
	return nil
}
