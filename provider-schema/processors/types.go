package processors

import (
	"fmt"
	"strings"

	"github.com/Azure/ms-terraform-lsp/provider-schema/azurerm/schema"
	"github.com/Azure/ms-terraform-lsp/provider-schema/processors/.tools/document-lint/model"
)

type TerraformObjects map[string]*TerraformObject

type TerraformObject struct {
	Name       string                             `json:"name"`
	Fields     map[string]*schema.SchemaAttribute `json:"fields"`
	ExampleHCL string                             `json:"example_hcl"`
	Timeouts   *model.Timeouts
	Import     model.Import
	Details    string `json:"details"` // Start from first h2 header (after description)
}

func (b *TerraformObject) GetName() string {
	return strings.TrimPrefix(b.Name, schema.InputDataSourcePrefix)
}

func (b *TerraformObject) IsDataSource() bool {
	return strings.HasPrefix(b.Name, schema.InputDataSourcePrefix)
}

func (b *TerraformObject) IsModule() bool {
	return strings.HasPrefix(b.Name, schema.AVMPrefix)
}

func (b *TerraformObject) GetResourceOrDataSourceDocLink() string {
	objectDocName, _ := strings.CutPrefix(b.GetName(), schema.AzureRMPrefix)
	if b.IsDataSource() {
		return fmt.Sprintf(schema.TerraformDocsURL, schema.DefaultDocVersion, schema.DataSources, objectDocName)
	}
	return fmt.Sprintf(schema.TerraformDocsURL, schema.DefaultDocVersion, schema.Resources, objectDocName)
}

func (b *TerraformObject) GetModuleDocLink() string {
	moduleName := strings.Split(b.Name, "/")[1]
	return fmt.Sprintf(schema.AVMDocsURL, moduleName, schema.DefaultDocVersion)
}

func (b *TerraformObject) GetModuleGitHubIssueLink() string {
	moduleName := strings.Split(b.Name, "/")[1]
	return fmt.Sprintf(schema.AVMGitHubIssuesURL, moduleName)
}

func (b *TerraformObject) GetRaiseAVMGitHubIssueLink() string {
	moduleName := strings.Split(b.Name, "/")[1]
	return fmt.Sprintf(schema.AVMNewGitHubIssuesURL, moduleName)
}

func (b *TerraformObject) GetGitHubIssueLink() string {
	return fmt.Sprintf(schema.GitHubIssuesURL, b.GetName())
}

func (b *TerraformObject) GetRaiseGitHubIssueLink() string {
	return fmt.Sprintf(schema.NewGitHubIssuesURL, fmt.Sprintf("`%s`", b.GetName()))
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

func (b *TerraformObject) GetDocContent() string {
	return b.Details
}
