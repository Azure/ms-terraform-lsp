package processors

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azurerm-lsp/provider-schema/processors/.tools/document-lint/md"
	"github.com/Azure/azurerm-lsp/provider-schema/processors/.tools/document-lint/model"
)

// ProcessMarkdown processes all markdown files in the given directory and returns structured output
func ProcessMarkdown(providerDir string) (map[string]*model.ResourceDoc, error) {
	allDocs := make(map[string]*model.ResourceDoc)

	// do below for resources and data sources
	resourceMarkdownDir := filepath.Join(providerDir, "website", "docs", "r")
	dataSourceMarkdownDir := filepath.Join(providerDir, "website", "docs", "d")

	// Walk through the markdown directory
	processDir := func(dir string, isDataSource bool) error {
		err := filepath.Walk(providerDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".markdown" {
				// Parse the markdown file
				mark := md.MustNewMarkFromFile(path)
				doc := mark.BuildResourceDoc()

				if isDataSource {
					doc.ResourceName = "datasource#" + doc.ResourceName
				}

				allDocs[doc.ResourceName] = doc
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("error walking through markdown directory: %v", err)
		}

		return nil
	}

	err := processDir(resourceMarkdownDir, false)
	if err != nil {
		return nil, fmt.Errorf("error processing resource markdown files: %v", err)
	}

	err = processDir(dataSourceMarkdownDir, true)
	if err != nil {
		return nil, fmt.Errorf("error processing data source markdown files: %v", err)
	}

	return allDocs, nil
}
