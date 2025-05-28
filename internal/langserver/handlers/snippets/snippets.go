package snippets

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	lsp "github.com/microsoft/msgraph-lsp/internal/protocol"
)

type Snippet struct {
	ResourceType string
	Fields       []Field
}

type Field struct {
	Name  string
	Value string
}

func (field Field) Order() int {
	switch field.Name {
	case "url":
		return -1
	case "parent_id":
		return 0
	case "body":
		return 1
	case "response_export_values":
		return 2
	default:
		return 3
	}
}

func CodeSampleCandidates(block *hclsyntax.Block, editRange lsp.Range, data []byte) []lsp.CompletionItem {
	if block == nil || block.Type == "data" {
		return nil
	}

	if len(block.Labels) != 2 || block.Labels[0] != "msgraph_resource" {
		return nil
	}

	urlValue := ""
	for _, attr := range block.Body.Attributes {
		if attr.Name == "url" {
			urlValue = strings.Trim(stringValue(data, attr.Expr.Range()), `"`)
			break
		}
	}

	resourceType := parseResourceType(urlValue)
	if snippet, ok := snippetMap[strings.ToLower(resourceType)]; ok {
		newText := ""
		for _, field := range snippet.Fields {
			if _, ok := block.Body.Attributes[field.Name]; ok {
				continue
			}
			newText += field.Value + "\n"
		}
		if newText == "" {
			return nil
		}

		event := lsp.TelemetryEvent{
			Version: lsp.TelemetryFormatVersion,
			Name:    "textDocument/completion",
			Properties: map[string]interface{}{
				"kind": "code-sample",
				"type": urlValue,
			},
		}
		data, _ := json.Marshal(event)

		return []lsp.CompletionItem{
			{
				Label:  "code sample",
				Kind:   lsp.SnippetCompletion,
				Detail: "Code Sample",
				Documentation: lsp.MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("```\n%s\n```\n", newText),
				},
				SortText:         "0",
				InsertTextFormat: lsp.SnippetTextFormat,
				InsertTextMode:   lsp.AdjustIndentation,
				TextEdit: &lsp.TextEdit{
					Range:   editRange,
					NewText: newText,
				},
				Command: &lsp.Command{
					Title:     "",
					Command:   "msgraph.telemetry",
					Arguments: []json.RawMessage{data},
				},
			},
		}
	}

	return nil
}

func parseResourceType(typeValue string) string {
	parts := strings.Split(typeValue, "/")
	for i, part := range parts {
		if strings.HasSuffix(part, "}") {
			parts[i] = "{}"
		}
	}
	return strings.Join(parts, "/")
}
