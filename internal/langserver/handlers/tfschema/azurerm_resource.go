package tfschema

import (
	"fmt"
	"strings"

	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	provider_schema "github.com/Azure/ms-terraform-lsp/provider-schema"
	"github.com/Azure/ms-terraform-lsp/provider-schema/azurerm/schema"
	"github.com/zclconf/go-cty/cty"
)

var _ Resource = &AzureRMResource{}

type AzureRMResource struct {
}

func (a AzureRMResource) ResourceDocumentation(resourceType string) string {
	parts := strings.Split(resourceType, ".")
	if len(parts) < 2 {
		return ""
	}

	content, _, err := provider_schema.GetResourceContent(parts[1])
	if err != nil {
		return ""
	}
	return content
}

func (a AzureRMResource) ListProperties(blockPath string) []Property {
	parts := strings.Split(blockPath, ".")
	if len(parts) < 2 {
		return nil
	}

	resourceName := parts[1]
	path := strings.Join(parts[2:], ".")

	props, err := provider_schema.ListDirectProperties(resourceName, path)
	if err != nil {
		return nil
	}

	var items []Property
	for _, p := range props {
		content, prop, err := provider_schema.GetAttributeContent(resourceName, p.AttributePath)
		if err != nil || prop == nil {
			continue
		}
		items = append(items, ToProperty(p, content))
	}
	return items
}

func (a AzureRMResource) GetProperty(propertyPath string) *Property {
	parts := strings.Split(propertyPath, ".")
	if len(parts) < 2 {
		return nil
	}
	resourceName := parts[1]
	path := strings.Join(parts[2:], ".")

	values, _ := provider_schema.GetPossibleValuesForProperty(resourceName, path)
	content, prop, _ := provider_schema.GetAttributeContent(resourceName, path)

	fixedItems := make([]lsp.CompletionItem, 0)
	for _, val := range values {
		fixedItems = append(fixedItems, lsp.CompletionItem{
			Label:  fmt.Sprintf(`"%s"`, val),
			Kind:   lsp.ValueCompletion,
			Detail: fmt.Sprintf("Possible value for %s", path),
			Documentation: lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: content,
			},
			TextEdit: &lsp.TextEdit{
				NewText: fmt.Sprintf(`"%s"`, val),
			},
		})
	}

	out := &Property{}
	if prop != nil {
		property := ToProperty(prop, content)
		out = &property
	}
	out.MarkdownDescription = content
	out.ValueCandidatesFunc = FixedValueCandidatesFunc(fixedItems)
	return out
}

func (a AzureRMResource) Match(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return false
	}
	return strings.HasPrefix(parts[1], "azurerm_") && parts[0] == "resource"
}

func ToProperty(p *schema.SchemaAttribute, content string) Property {
	insertText := p.Name
	propType := ""
	if p.AttributeType.IsPrimitiveType() {
		switch p.AttributeType {
		case cty.String:
			insertText = fmt.Sprintf(`%s = "$0"`, p.Name)
			propType = "string"
		case cty.Bool:
			insertText = fmt.Sprintf(`%s = $0`, p.Name)
			propType = "bool"
		case cty.Number:
			insertText = fmt.Sprintf(`%s = $0`, p.Name)
			propType = "number"
		default:
			insertText = fmt.Sprintf(`%s = $0`, p.Name)
			propType = "object"
		}
	} else if p.AttributeType.IsMapType() || p.AttributeType.IsObjectType() {
		// invalid nesting mode
		if p.NestingMode == 0 {
			insertText = fmt.Sprintf(`%s = { $0 }`, p.Name)
		} else {
			insertText = fmt.Sprintf(`%s {$0}`, p.Name)
		}
		propType = "object"
	} else if p.AttributeType.IsListType() || p.AttributeType.IsSetType() {
		insertText = fmt.Sprintf(`%s = [$0]`, p.Name)
		propType = "list"
	}

	modifier := "Optional"
	if p.Required {
		modifier = "Required"
	}
	return Property{
		Name:                  p.Name,
		Modifier:              modifier,
		Type:                  propType,
		MarkdownDescription:   content,
		CompletionNewText:     insertText,
		GenericCandidatesFunc: nil,
		ValueCandidatesFunc:   nil,
		NestedProperties:      nil,
	}
}
