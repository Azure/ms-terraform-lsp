package tfschema

import (
	"fmt"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	provider_schema "github.com/Azure/ms-terraform-lsp/provider-schema"
	"strings"
)

var _ Resource = &AVMModule{}

type AVMModule struct {
}

// GetProperty input: module.Azure/avm-ptn-aiml-ai-foundry/azurerm
func (m AVMModule) GetProperty(propertyName string) *Property {
	parts := strings.Split(propertyName, ".")
	if len(parts) < 2 {
		return nil
	}

	moduleName := parts[1]
	path := strings.Join(parts[2:], ".")

	// It's OK for AVM modules to not have possible values.
	values, _ := provider_schema.GetPossibleValuesForProperty(moduleName, path, false)
	content, prop, err := provider_schema.GetModuleAttributeContent(moduleName, path)
	if err != nil || prop == nil {
		return nil
	}

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

	property := ToProperty(prop, content)
	out := &property
	out.MarkdownDescription = content
	out.ValueCandidatesFunc = FixedValueCandidatesFunc(fixedItems)
	return out
}

func (m AVMModule) Match(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return false
	}

	return parts[0] == "module"
}

func (m AVMModule) ListProperties(blockPath string) []Property {
	parts := strings.Split(blockPath, ".")
	if len(parts) < 2 {
		return nil
	}

	moduleName := parts[1]
	path := strings.Join(parts[2:], ".")

	props, err := provider_schema.ListDirectProperties(moduleName, path, false)
	if err != nil {
		return nil
	}

	var items []Property
	for _, p := range props {
		content, prop, err := provider_schema.GetModuleAttributeContent(moduleName, p.AttributePath)
		if err != nil || prop == nil {
			continue
		}
		items = append(items, ToProperty(p, content))
	}
	return items
}

func (m AVMModule) ResourceDocumentation(resourceType string) string {
	parts := strings.Split(resourceType, ".")
	if len(parts) < 2 {
		return ""
	}

	content, err := provider_schema.GetModuleContent(parts[1])
	if err != nil {
		return ""
	}
	return content
}
