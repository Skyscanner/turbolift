package init

import (
	_ "embed"
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

	files := map[string]string{
		".gitignore": gitignoreTemplate,
		".turbolift": turboliftTemplate,
		"README.md":  readmeTemplate,
		"repos.txt":  reposTemplate,
	}
	for filename, templateFile := range files {
		err := applyTemplate(filepath.Join(campaignName, filename), templateFile, data)
		if err != nil {
			c.Printf(colors.Red("Error when templating file: %s\n"), err)
			return
		}
	}

	c.Println(colors.Green("✅ turbolift init is done - next:"))
	c.Println("1. Run", colors.Cyan("cd ", campaignName))
	c.Println("2. Update repos.txt with the names of the repos that need changing (either manually or using a tool to generate a list of repos)")
	c.Println("3. Run", colors.Cyan("turbolift clone"))
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
