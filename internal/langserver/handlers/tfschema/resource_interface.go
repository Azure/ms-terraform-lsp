package tfschema

import (
	"fmt"

	ilsp "github.com/Azure/ms-terraform-lsp/internal/lsp"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Resource interface {
	// GetProperty input: resource.azurerm_data_factory.identity.type
	GetProperty(propertyName string) *Property

	// Match input: resource.azurerm_data_factory
	Match(name string) bool

	// ListProperties input: resource.azurerm_data_factory.identity
	ListProperties(blockPath string) []Property

	// ResourceDocumentation input: msgraph_resource.applications
	ResourceDocumentation(resourceType string) string
}

type Property struct {
	Name                string
	Modifier            string
	Type                string
	Description         string
	MarkdownDescription string
	CompletionNewText   string

	GenericCandidatesFunc GenericCandidatesFunc
	ValueCandidatesFunc   ValueCandidatesFunc
	CustomizedHoverFunc   CustomizedHoverFunc
	NestedProperties      []Property
}

func (property *Property) ToHover(r hcl.Range) *lsp.Hover {
	if property == nil {
		return nil
	}
	content := property.MarkdownDescription
	if content == "" {
		content = fmt.Sprintf("```\n%s: %s(%s)\n```\n%s", property.Name, property.Modifier, property.Type, property.Description)
	}
	return &lsp.Hover{
		Range: ilsp.HCLRangeToLSP(r),
		Contents: lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: content,
		},
	}
}

type GenericCandidatesFunc func(data []byte, filename string, block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, property *Property) []lsp.CompletionItem
type ValueCandidatesFunc func(prefix *string, r lsp.Range) []lsp.CompletionItem
type CustomizedHoverFunc func(block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, data []byte) *lsp.Hover
