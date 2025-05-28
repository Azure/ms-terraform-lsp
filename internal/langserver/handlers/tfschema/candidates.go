package tfschema

import (
	"fmt"
	"strings"

	"github.com/Azure/azurerm-lsp/internal/msgraph"
	"github.com/Azure/azurerm-lsp/internal/parser"
	lsp "github.com/Azure/azurerm-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/ms-henglu/go-msgraph-types/types"
)

func PropertiesCandidates(props []Property, r *lsp.Range) []lsp.CompletionItem {
	candidates := make([]lsp.CompletionItem, 0)

	for index, prop := range props {
		documentation := fmt.Sprintf("Type: `%s`  \n%s\n", prop.Type, prop.Description)
		if prop.MarkdownDescription != "" {
			documentation = prop.MarkdownDescription
		}

		completionItem := lsp.CompletionItem{
			Label:  prop.Name,
			Kind:   lsp.PropertyCompletion,
			Detail: fmt.Sprintf("%s (%s)", prop.Name, prop.Modifier),
			Documentation: lsp.MarkupContent{
				Kind:  "markdown",
				Value: documentation,
			},
			SortText:         fmt.Sprintf("%04d", index),
			InsertTextFormat: lsp.SnippetTextFormat,
			InsertTextMode:   lsp.AdjustIndentation,
			Command:          constTriggerSuggestCommand(),
		}

		if r != nil {
			completionItem.TextEdit = &lsp.TextEdit{
				Range:   *r,
				NewText: prop.CompletionNewText,
			}
		} else {
			completionItem.InsertText = prop.CompletionNewText
		}

		candidates = append(candidates, completionItem)
	}
	return candidates
}

func valueCandidates(values []string, r lsp.Range, isOrdered bool) []lsp.CompletionItem {
	candidates := make([]lsp.CompletionItem, 0)
	for index, value := range values {
		literal := strings.Trim(value, `"`)
		sortText := "0" + literal
		if isOrdered {
			sortText = fmt.Sprintf("%04d", index)
		}
		candidates = append(candidates, lsp.CompletionItem{
			Label: value,
			Kind:  lsp.ValueCompletion,
			Documentation: lsp.MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("Value: `%s`  \n", literal),
			},
			SortText:         sortText,
			InsertTextFormat: lsp.PlainTextTextFormat,
			InsertTextMode:   lsp.AdjustIndentation,
			TextEdit: &lsp.TextEdit{
				Range:   r,
				NewText: value,
			},
		})
	}
	return candidates
}

func urlCandidates(_ []byte, _ string, block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, _ *Property) []lsp.CompletionItem {
	apiVersion := "v1.0"
	if v := parser.BlockAttributeLiteralValue(block, "api_version"); v != nil {
		apiVersion = *v
	}

	resources := make([]types.ResourceType, 0)
	switch block.Type {
	case "resource":
		resources = msgraph.SchemaLoader.ListResources(apiVersion)
	case "data":
		resources = msgraph.SchemaLoader.ListReadableResources(apiVersion)
	}
	candidates := make([]lsp.CompletionItem, 0)
	r := editRangeFromExprRange(attribute.Expr, pos)
	for _, resource := range resources {
		doc := fmt.Sprintf("Resource: `%s`  \nSummary: %s  \n", resource.Url, resource.Name)
		if resource.Description != "" {
			doc += fmt.Sprintf("Description: %s  \n", resource.Description)
		}
		if resource.ExternalDocs != nil {
			doc += fmt.Sprintf("External Docs: [%s](%s)  \n", resource.ExternalDocs.Description, resource.ExternalDocs.Url)
		}
		newText := fmt.Sprintf(`"%s"`, strings.TrimPrefix(resource.Url, "/"))
		newText = strings.ReplaceAll(newText, "$", "\\$")
		candidates = append(candidates, lsp.CompletionItem{
			Label: fmt.Sprintf(`"%s"`, strings.TrimPrefix(resource.Url, "/")),
			Kind:  lsp.ValueCompletion,
			Documentation: lsp.MarkupContent{
				Kind:  "markdown",
				Value: doc,
			},
			SortText:         resource.Url,
			InsertTextFormat: lsp.SnippetTextFormat,
			InsertTextMode:   lsp.AdjustIndentation,
			TextEdit: &lsp.TextEdit{
				Range:   r,
				NewText: newText,
			},
		})
	}
	return candidates
}

func dynamicPlaceholderCandidate() lsp.CompletionItem {
	return lsp.CompletionItem{
		Label: `{}`,
		Kind:  lsp.ValueCompletion,
		Documentation: lsp.MarkupContent{
			Kind:  "markdown",
			Value: "dynamic attribute allows any valid HCL object.",
		},
		SortText:         `{}`,
		InsertTextFormat: lsp.SnippetTextFormat,
		InsertTextMode:   lsp.AdjustIndentation,
		TextEdit: &lsp.TextEdit{
			NewText: "{\n\t$0\n}",
		},
		Command: constTriggerSuggestCommand(),
	}
}

func apiVersionCandidates(_ *string, r lsp.Range) []lsp.CompletionItem {
	return valueCandidates([]string{
		`"v1.0"`,
		`"beta"`,
	}, r, true)
}

func constTriggerSuggestCommand() *lsp.Command {
	return &lsp.Command{
		Command: "editor.action.triggerSuggest",
		Title:   "Suggest",
	}
}
