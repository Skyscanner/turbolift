package cmd

import (
	_ "embed"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"html/template"
	"log"
	"os"
	"path/filepath"
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

type TemplateVariables struct {
	CampaignName string
}

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
	// Create a directory for both the campaign and its nested work directory
	workDirectory := filepath.Join(campaignName, "work")
	err := os.MkdirAll(workDirectory, os.ModeDir|0755)

	if err != nil {
		log.Panic("Unable to create directory ", workDirectory, ": ", err)
	}

	data := TemplateVariables{
		CampaignName: campaignName,
	}

	applyTemplate(filepath.Join(campaignName, ".gitignore"), gitignoreTemplate, data)
	applyTemplate(filepath.Join(campaignName, ".turbolift"), turboliftTemplate, data)
	applyTemplate(filepath.Join(campaignName, "README.md"), readmeTemplate, data)
	applyTemplate(filepath.Join(campaignName, "repos.txt"), reposTemplate, data)

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Println(green("✅ turbolift init is done - next:"))
	fmt.Println("1. Run", cyan("cd ", campaignName))
	fmt.Println("2. Update repos.txt with the names of the repos that need changing (either manually or using a tool to generate a list of repos)")
	fmt.Println("3. Run", cyan("turbolift clone"))
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
