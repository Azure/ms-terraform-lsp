package processors

import (
	"fmt"
	"strings"

	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
	"github.com/Azure/azurerm-lsp/provider-schema/processors/.tools/document-lint/model"
)

type TerraformObjects map[string]*TerraformObject

type TerraformObject struct {
	Name       string                             `json:"name"`
	Fields     map[string]*schema.SchemaAttribute `json:"fields"`
	ExampleHCL string                             `json:"example_hcl"`
	Timeouts   *model.Timeouts                    // nil if no timeouts part in document
	Import     model.Import
}

func (b *TerraformObject) GetName() string {
	return strings.TrimPrefix(b.Name, schema.DataSourcePrefix)
}

func (b *TerraformObject) IsDataSource() bool {
	return strings.HasPrefix(b.Name, schema.DataSourcePrefix)
}

func (b *TerraformObject) GetResourceOrDataSourceDocLink() string {
	if b.IsDataSource() {
		return fmt.Sprintf(schema.TerraformDocsURL, schema.DefaultDocVersion, schema.DataSources, b.GetName())
	}
	return fmt.Sprintf(schema.TerraformDocsURL, schema.DefaultDocVersion, schema.Resources, b.GetName())
}

func (b *TerraformObject) GetSnippet() string {
	snippet := strings.TrimSpace(b.ExampleHCL)
	snippet = strings.TrimPrefix(snippet, "```hcl")
	snippet = strings.TrimPrefix(snippet, "\n")
	snippet = strings.TrimSuffix(snippet, "\n")
	snippet = strings.TrimSuffix(snippet, "```")
	snippet = strings.TrimSuffix(snippet, "\n")
	return strings.TrimSpace(snippet)
}
