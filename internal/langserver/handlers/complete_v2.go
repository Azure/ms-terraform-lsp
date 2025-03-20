package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	lsctx "github.com/Azure/azurerm-lsp/internal/context"
	ilsp "github.com/Azure/azurerm-lsp/internal/lsp"
	"github.com/Azure/azurerm-lsp/internal/parser"
	lsp "github.com/Azure/azurerm-lsp/internal/protocol"
	"github.com/Azure/azurerm-lsp/internal/utils"
	provider_schema "github.com/Azure/azurerm-lsp/provider-schema"
	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (svc *service) HandleComplete(ctx context.Context, params lsp.CompletionParams) ([]lsp.CompletionItem, error) {
	docContent, docFileName, err := GetDocumentContent(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %v", err)
	}

	docContents := strings.Split(docContent, "\n")
	lineContent, err := getCurrentLineContent(docContents, params.TextDocumentPositionParams.Position.Line)
	if err != nil {
		return nil, fmt.Errorf("failed to get current line content: %v", err)
	}

	file, diags := hclsyntax.ParseConfig([]byte(docContent), docFileName, hcl.InitialPos)
	if diags != nil && diags.HasErrors() {
		if utils.MatchAnyPrefix(lineContent, "azurerm") {
			return suggestResourcesAndDataSources(params), nil
		}

		return nil, fmt.Errorf("failed to parse document: %v", diags)
	}

	hclPos := ilsp.LSPPosToHCL(params.TextDocumentPositionParams.Position)
	body, isHcl := file.Body.(*hclsyntax.Body)
	if !isHcl {
		return nil, fmt.Errorf("file is not HCL")
	}

	block := parser.BlockAtPos(body, hclPos)
	if block == nil || len(block.Labels) == 0 || !strings.HasPrefix(block.Labels[0], "azurerm") {
		return suggestResourcesAndDataSources(params), nil
	}

	resourceName := block.Labels[0]
	attribute := parser.AttributeAtPos(block, hclPos)
	subBlock := parser.BlockAtPos(block.Body, hclPos)

	svc.logger.Printf("resourceName: %s, attribute: %v, subBlock: %v", jsonLog(resourceName), jsonLog(attribute), jsonLog(subBlock))

	switch {
	case attribute != nil:
		path := buildNestedPath(block, attribute)
		return completeAttribute(resourceName, path)
	case subBlock != nil:
		path := buildNestedPath(block, subBlock)
		return completeNestedBlock(resourceName, path)
	default:
		return completeTopLevelProperties(resourceName)
	}
}

func jsonLog(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func getCurrentLineContent(contents []string, line uint32) (string, error) {
	if line < 0 || int(line) >= len(contents) {
		return "", fmt.Errorf("invalid line number")
	}

	return strings.TrimSpace(contents[line]), nil
}

// buildNestedPath constructs the path to a nested block or attribute by traversing from the top-level resource block.
func buildNestedPath(topLevelBlock *hclsyntax.Block, targetNode hclsyntax.Node) string {
	var pathParts []string

	// Helper function to recursively traverse the AST
	var traverse func(block *hclsyntax.Block, target hclsyntax.Node) bool
	traverse = func(block *hclsyntax.Block, target hclsyntax.Node) bool {
		// Check if the target is within this block's body
		for _, attr := range block.Body.Attributes {
			if attr == target {
				pathParts = append(pathParts, attr.Name)
				return true
			}
		}

		// Check nested blocks
		for _, nestedBlock := range block.Body.Blocks {
			if nestedBlock == target {
				pathParts = append(pathParts, nestedBlock.Type)
				return true
			}

			// Recursively traverse nested blocks
			if traverse(nestedBlock, target) {
				pathParts = append([]string{nestedBlock.Type}, pathParts...)
				return true
			}
		}

		return false
	}

	// Start traversal from the top-level block
	if traverse(topLevelBlock, targetNode) {
		return strings.Join(pathParts, ".")
	}

	// If the target node is not found, return an empty path
	return ""
}

// completeAttribute handles completion for attributes
func completeAttribute(resourceName, path string) ([]lsp.CompletionItem, error) {
	possibleValues, err := provider_schema.GetPossibleValuesForProperty(resourceName, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get possible values: %v", err)
	}

	candidateList := make([]lsp.CompletionItem, 0, len(possibleValues))
	for _, possibleValue := range possibleValues {
		candidateList = append(candidateList, lsp.CompletionItem{
			Label:      "☁️(possible value) " + possibleValue,
			Kind:       lsp.SnippetCompletion,
			Detail:     "Code Snippet",
			InsertText: possibleValue,
		})
	}

	return candidateList, nil
}

// completeNestedBlock handles completion for nested blocks
func completeNestedBlock(resourceName, path string) ([]lsp.CompletionItem, error) {
	propertyInfo, err := provider_schema.GetPropertyInfo(resourceName, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get property info: %v", err)
	}

	candidateList := make([]lsp.CompletionItem, 0, len(propertyInfo.Fields))
	for nestedProperty, nestedPropertyInfo := range propertyInfo.Fields {
		candidateList = append(candidateList, createCompletionItemFromSchemaAttribute(nestedProperty, nestedPropertyInfo))
	}

	return candidateList, nil
}

// completeTopLevelProperties handles completion for top-level properties
func completeTopLevelProperties(resourceName string) ([]lsp.CompletionItem, error) {
	properties, err := provider_schema.ListDirectProperties(resourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to list properties: %v", err)
	}

	candidateList := make([]lsp.CompletionItem, 0, len(properties))
	for _, property := range properties {
		candidateList = append(candidateList, createCompletionItemFromSchemaAttribute(property.Name, property))
	}

	return candidateList, nil
}

func suggestResourcesAndDataSources(params lsp.CompletionParams) []lsp.CompletionItem {
	resources := provider_schema.ListAllResources()
	dataSources := provider_schema.ListAllDataSources()
	candidateList := make([]lsp.CompletionItem, 0, len(resources)+len(dataSources))

	lineRange := getCurrentLineRange(params)

	for _, resource := range resources {
		snippet, err := provider_schema.GetSnippet(resource)
		if err != nil {
			continue
		}

		candidateList = append(candidateList, lsp.CompletionItem{
			Label:            "☁️(resource) " + resource,
			InsertText:       snippet,
			InsertTextFormat: lsp.SnippetTextFormat,
			Kind:             lsp.SnippetCompletion,
			Detail:           "Resource",
			TextEdit: &lsp.TextEdit{
				Range:   lineRange,
				NewText: snippet,
			},
		})
	}

	for _, dataSource := range dataSources {
		snippet, err := provider_schema.GetSnippet(dataSource)
		if err != nil {
			continue
		}

		candidateList = append(candidateList, lsp.CompletionItem{
			Label:            "☁️(data source) " + dataSource,
			InsertText:       snippet,
			InsertTextFormat: lsp.SnippetTextFormat,
			Kind:             lsp.SnippetCompletion,
			Detail:           "Data Source",
			TextEdit: &lsp.TextEdit{
				Range:   lineRange,
				NewText: snippet,
			},
		})
	}

	return candidateList
}

func createCompletionItemFromSchemaAttribute(name string, attr *schema.SchemaAttribute) lsp.CompletionItem {
	label := fmt.Sprintf("☁️(property) %s (%s)", name, attr.AttributeType.FriendlyName())
	if attr.Required {
		label += " (required)"
	} else if attr.Optional {
		label += " (optional)"
	} else if attr.Computed {
		label += " (computed)"
	}

	if attr.Default != nil {
		label += fmt.Sprintf(" (default: %v)", attr.Default)
	}

	label += fmt.Sprintf(" (%s)", attr.Description)

	return lsp.CompletionItem{
		Label:      label,
		Kind:       lsp.SnippetCompletion,
		InsertText: name,
	}
}

func GetDocumentContent(ctx context.Context, documentURI lsp.DocumentURI) (string, string, error) {
	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return "", "", err
	}

	doc, err := fs.GetDocument(ilsp.FileHandlerFromDocumentURI(documentURI))
	if err != nil {
		return "", "", fmt.Errorf("failed to get document: %v", err)
	}

	data, err := doc.Text()
	if err != nil {
		return "", "", fmt.Errorf("failed to get document text: %v", err)
	}

	return string(data), doc.Filename(), nil
}

func getCurrentLineRange(params lsp.CompletionParams) lsp.Range {
	line := params.TextDocumentPositionParams.Position.Line
	return lsp.Range{
		Start: lsp.Position{Line: line, Character: 0},
		End:   lsp.Position{Line: line, Character: 1000}, // Use a large number to cover the entire line
	}
}
