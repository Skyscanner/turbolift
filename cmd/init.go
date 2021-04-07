package cmd

import (
	"github.com/spf13/cobra"
	"html/template"
	"log"
	"os"
	"path/filepath"

	_ "embed"
)

var (
	initCmd = createInitCmd()

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

func init() {
	rootCmd.AddCommand(initCmd)
}

func createInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Turbolift campaign directory",
		Run:   run,
	}

	cmd.Flags().StringVarP(&campaignName, "name", "n", "", "Campaign name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func run(*cobra.Command, []string) {
	err := os.Mkdir(campaignName, os.ModeDir|0755)

	if err != nil {
		log.Panic("Unable to create workspace directory ", campaignName, ": ", err)
	}

	type TemplateVariables struct {
		CampaignName string
	}

	data := TemplateVariables{
		CampaignName: campaignName,
	}

	applyTemplate(filepath.Join(campaignName, ".gitignore"), gitignoreTemplate, data)
	applyTemplate(filepath.Join(campaignName, ".turbolift"), turboliftTemplate, data)
	applyTemplate(filepath.Join(campaignName, "README.md"), readmeTemplate, data)
	applyTemplate(filepath.Join(campaignName, "repos.txt"), reposTemplate, data)
}

// Applies a given template and data to produce a file with the outputFilename
func applyTemplate(outputFilename string, templateContent string, data interface{}) {
	readme, err := os.Create(outputFilename)

	parsedTemplate, err := template.New("").Parse(templateContent)

	if err != nil {
		log.Panic("Unable to parse template")
	}

	err = parsedTemplate.Execute(readme, data)

	if err != nil {
		log.Panic("Unable to write templated file")
	}
}
