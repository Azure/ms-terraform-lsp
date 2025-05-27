package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azurerm-lsp/internal/parser"
	"github.com/Azure/azurerm-lsp/internal/protocol"
	"github.com/Azure/azurerm-lsp/internal/utils"
	"github.com/Azure/azurerm-lsp/provider-schema"
	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
	"github.com/zclconf/go-cty/cty"
)

func (svc *service) HandleComplete(ctx context.Context, params protocol.CompletionParams) (protocol.CompletionList, error) {
	var list protocol.CompletionList

	docContent, docFileName, err := parser.GetDocumentContent(ctx, params.TextDocument.URI)
	if err != nil {
		return list, err
	}

	if shouldGiveTopLevelCompletions(docContent, int(params.Position.Line)) {
		list.Items = GetTopLevelCompletions(params)
		return list, nil
	}

	ctxInfo, diags, err := parser.BuildHCLContext(docContent, docFileName, params.Position)
	if err != nil || (diags != nil && diags.HasErrors()) {
		docContent, fieldName, _, err := parser.AttemptReparse(docContent, params.Position.Line)
		if err != nil {
			return list, nil
		}

		ctxInfo, diags, err = parser.BuildHCLContext(docContent, docFileName, params.Position)
		if err != nil || (diags != nil && diags.HasErrors()) {
			return list, nil
		}

		if ctxInfo.Block != nil || ctxInfo.SubBlock != nil || ctxInfo.Attribute != nil {
			if ctxInfo.ParsedPath != "" {
				fieldName = ctxInfo.ParsedPath + "." + fieldName
			}

			list.Items = GetAttributeCompletions(ctxInfo.Resource, fieldName)
			return list, nil
		}

		return list, nil
	}

	switch {
	case ctxInfo.Attribute != nil:
		list.Items = GetAttributeCompletions(ctxInfo.Resource, ctxInfo.ParsedPath)
		return list, nil
	case ctxInfo.SubBlock != nil || ctxInfo.Block != nil:
		list.Items = GetBlockAttributeCompletions(ctxInfo.Resource, ctxInfo.ParsedPath)
		return list, nil
	}

	return list, nil
}

func GetTopLevelCompletions(params protocol.CompletionParams) []protocol.CompletionItem {
	resources := provider_schema.ListAllResources()
	dataSources := provider_schema.ListAllDataSources()
	lineRange := getLineRange(params)

	var items []protocol.CompletionItem
	for _, name := range append(resources, dataSources...) {
		snippet, err := provider_schema.GetSnippet(name)
		if err != nil {
			continue
		}

		content, isDataSource, err := provider_schema.GetResourceContent(name)
		if err != nil {
			continue
		}

		kind := "resource"
		if isDataSource {
			kind = "data source"
		}

		items = append(items, protocol.CompletionItem{
			Label:            name,
			InsertText:       snippet,
			InsertTextFormat: protocol.SnippetTextFormat,
			Kind:             protocol.SnippetCompletion,
			Detail:           "AzureRM " + kind,
			TextEdit: &protocol.TextEdit{
				Range:   lineRange,
				NewText: snippet,
			},
			Documentation: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: content,
			},
		})
	}
	return items
}

func GetBlockAttributeCompletions(resourceName, path string) []protocol.CompletionItem {
	props, err := provider_schema.ListDirectProperties(resourceName, path)
	if err != nil {
		return nil
	}

	var items []protocol.CompletionItem
	for _, p := range props {
		content, prop, err := provider_schema.GetAttributeContent(resourceName, p.AttributePath)
		if err != nil || prop == nil {
			continue
		}

		insertText := p.Name
		if p.AttributeType.IsPrimitiveType() {
			switch p.AttributeType {
			case cty.String:
				insertText = fmt.Sprintf(`%s = "$0"`, p.Name)
			case cty.Bool:
				insertText = fmt.Sprintf(`%s = $0`, p.Name)
			case cty.Number:
				insertText = fmt.Sprintf(`%s = $0`, p.Name)
			default:
				insertText = fmt.Sprintf(`%s = $0`, p.Name)
			}
		} else if p.AttributeType.IsMapType() || p.AttributeType.IsObjectType() {
			// invalid nesting mode
			if p.NestingMode == 0 {
				insertText = fmt.Sprintf(`%s = { $0 }`, p.Name)
			} else {
				insertText = fmt.Sprintf(`%s {$0}`, p.Name)
			}
		} else if p.AttributeType.IsListType() || p.AttributeType.IsSetType() {
			insertText = fmt.Sprintf(`%s = [$0]`, p.Name)
		}

		items = append(items, protocol.CompletionItem{
			Label:            p.Name,
			Kind:             protocol.PropertyCompletion,
			SortText:         p.GetSortOrder(),
			InsertText:       insertText,
			InsertTextFormat: protocol.SnippetTextFormat,
			Documentation: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: content,
			},
			// Add this command to trigger suggestions after insertion
			Command: &protocol.Command{
				Title:   "Trigger Suggest",
				Command: "editor.action.triggerSuggest",
			},
		})
	}
	return items
}

func GetAttributeCompletions(resourceName, path string) []protocol.CompletionItem {
	values, err := provider_schema.GetPossibleValuesForProperty(resourceName, path)
	if err != nil {
		return nil
	}
	content, _, err := provider_schema.GetAttributeContent(resourceName, path)
	if err != nil {
		return nil
	}

	items := make([]protocol.CompletionItem, 0, len(values))
	for _, val := range values {
		items = append(items, protocol.CompletionItem{
			Label:  val,
			Kind:   protocol.ValueCompletion,
			Detail: "Possible value for " + path,
			Documentation: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: content,
			},
		})
	}

	return items
}

func getLineRange(params protocol.CompletionParams) protocol.Range {
	start := protocol.Position{Line: params.Position.Line, Character: 0}
	end := params.Position
	return protocol.Range{Start: start, End: end}
}

func shouldGiveTopLevelCompletions(content string, line int) bool {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return false
	}

	currentLine := strings.TrimSpace(lines[line])
	if !utils.MatchAnyPrefix(currentLine, schema.AzureRMPrefix) {
		return false
	}

	// Check if we're at root level by counting unclosed blocks
	openBlocks := 0
	for i := 0; i <= line; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.Contains(trimmed, "{") {
			openBlocks++
		}
		if strings.Contains(trimmed, "}") {
			openBlocks--
		}
	}

	return openBlocks == 0
}
