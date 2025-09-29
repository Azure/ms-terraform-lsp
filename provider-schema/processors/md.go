package processors

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/ms-terraform-lsp/provider-schema/azurerm/schema"
	"github.com/Azure/ms-terraform-lsp/provider-schema/processors/.tools/document-lint/md"
	"github.com/Azure/ms-terraform-lsp/provider-schema/processors/.tools/document-lint/model"
)

// ProcessMarkdown processes all markdown files in the given directory and returns structured output
func ProcessMarkdown(providerDir string) (map[string]*model.ResourceDoc, error) {
	allDocs := make(map[string]*model.ResourceDoc)

	// Convert to absolute path to ensure MustNewMarkFromFile works correctly
	absProviderDir, err := filepath.Abs(providerDir)
	if err != nil {
		return nil, fmt.Errorf("error converting provider directory to absolute path: %v", err)
	}

	// do below for resources and data sources
	resourceMarkdownDir := filepath.Join(absProviderDir, "website", "docs", "r")
	dataSourceMarkdownDir := filepath.Join(absProviderDir, "website", "docs", "d")

	// Walk through the markdown directory
	processDir := func(dir string, isDataSource bool) error {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".markdown" {
				// Parse the markdown file
				mark := md.MustNewMarkFromFile(path)
				if mark == nil {
					// Skip files that couldn't be parsed (e.g., non-absolute paths or read errors)
					fmt.Printf("Warning: Could not parse markdown file: %s\n", path)
					return nil
				}
				doc := mark.BuildResourceDoc()

				if isDataSource {
					doc.ResourceName = schema.InputDataSourcePrefix + doc.ResourceName
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

	err = processDir(resourceMarkdownDir, false)
	if err != nil {
		return nil, fmt.Errorf("error processing resource markdown files: %v", err)
	}

	err = processDir(dataSourceMarkdownDir, true)
	if err != nil {
		return nil, fmt.Errorf("error processing data source markdown files: %v", err)
	}

	return allDocs, nil
}
