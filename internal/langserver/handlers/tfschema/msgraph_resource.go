package tfschema

import (
	"fmt"
	"strings"

	"github.com/Azure/ms-terraform-lsp/internal/langserver/schema"
	ilsp "github.com/Azure/ms-terraform-lsp/internal/lsp"
	"github.com/Azure/ms-terraform-lsp/internal/msgraph"
	"github.com/Azure/ms-terraform-lsp/internal/parser"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/ms-henglu/go-msgraph-types/types"
)

var _ Resource = &MSGraphResource{}

type MSGraphResource struct {
	Name       string
	Properties []Property
}

func (r *MSGraphResource) ResourceDocumentation(resourceType string) string {
	parts := strings.Split(resourceType, "@")
	if len(parts) != 2 {
		return ""
	}
	apiVersion := parts[1]
	urlValue := parts[0]
	resourceDef := msgraph.SchemaLoader.GetResourceDefinition(apiVersion, urlValue)
	doc := fmt.Sprintf("Url: '%s'  \nSummary: %s  \nDescription: %s  \n", resourceDef.Url, resourceDef.Name, resourceDef.Description)
	if resourceDef.ExternalDocs != nil {
		doc = fmt.Sprintf("%s\n[%s](%s)", doc, resourceDef.ExternalDocs.Description, resourceDef.ExternalDocs.Url)
	}
	return doc
}

func (r *MSGraphResource) ListProperties(blockPath string) []Property {
	p := r.GetProperty(blockPath)
	if p == nil {
		return nil
	}
	return p.NestedProperties
}

func (r *MSGraphResource) Match(name string) bool {
	if r == nil {
		return false
	}
	return r.Name == name
}

func (r *MSGraphResource) GetProperty(propertyPath string) *Property {
	if r == nil {
		return nil
	}
	parts := strings.Split(propertyPath, ".")
	if len(parts) <= 2 {
		return nil
	}

	p := Property{
		NestedProperties: r.Properties,
	}
	if parts[2] == "" {
		return &p
	}
	for index := 2; index < len(parts); index++ {
		found := false
		for _, prop := range p.NestedProperties {
			if prop.Name == parts[index] {
				p = prop
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return &p
}

func msgraphBodyCandidates(data []byte, filename string, block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, property *Property) []lsp.CompletionItem {
	if attribute.Expr != nil {
		if _, ok := attribute.Expr.(*hclsyntax.LiteralValueExpr); ok && parser.ToLiteral(attribute.Expr) == nil {
			if property != nil {
				return property.ValueCandidatesFunc(nil, editRangeFromExprRange(attribute.Expr, pos))
			}
		}
	}

	urlValue := parser.ExtractMSGraphUrl(block, data)
	apiVersion := "v1.0"
	if v := parser.BlockAttributeLiteralValue(block, "api_version"); v != nil {
		apiVersion = *v
	}
	bodyDef := msgraph.SchemaLoader.GetResourceDefinition(apiVersion, urlValue)

	if bodyDef == nil {
		return nil
	}

	tokens, _ := hclsyntax.LexExpression(data[attribute.Expr.Range().Start.Byte:attribute.Expr.Range().End.Byte], filename, attribute.Expr.Range().Start)
	hclNode := parser.BuildHclNode(tokens)
	if hclNode == nil {
		return nil
	}

	return buildMSGraphBodyCandidates(hclNode, filename, pos, bodyDef)
}

func buildMSGraphBodyCandidates(hclNode *parser.HclNode, filename string, pos hcl.Pos, def types.TypeBase) []lsp.CompletionItem {
	candidateList := make([]lsp.CompletionItem, 0)
	hclNodes := parser.HclNodeArraysOfPos(hclNode, pos)
	if len(hclNodes) == 0 {
		return nil
	}
	lastHclNode := hclNodes[len(hclNodes)-1]

	switch {
	case parser.ContainsPos(lastHclNode.KeyRange, pos):
		// input a property with a prefix
		hclNodes := hclNodes[0 : len(hclNodes)-1]
		defs := schema.GetMSGraphDef(def.AsTypeBase(), hclNodes, 0)
		keys := make([]schema.Property, 0)
		for _, def := range defs {
			keys = append(keys, schema.GetMSGraphAllowedProperties(def)...)
		}
		editRange := ilsp.HCLRangeToLSP(lastHclNode.KeyRange)
		candidateList = keyCandidates(keys, editRange, lastHclNode)
	case !lastHclNode.KeyRange.Empty() && !lastHclNode.EqualRange.Empty() && lastHclNode.Children == nil:
		// input property =
		defs := schema.GetMSGraphDef(def.AsTypeBase(), hclNodes, 0)
		values := make([]string, 0)
		for _, def := range defs {
			values = append(values, schema.GetMSGraphAllowedValues(def)...)
		}
		editRange := lastHclNode.ValueRange
		if lastHclNode.Value == nil {
			editRange.End = pos
		}
		candidateList = valueCandidates(values, ilsp.HCLRangeToLSP(editRange), false)
	case parser.ContainsPos(lastHclNode.ValueRange, pos):
		// input a property
		defs := schema.GetMSGraphDef(def.AsTypeBase(), hclNodes, 0)
		keys := make([]schema.Property, 0)
		for _, def := range defs {
			keys = append(keys, schema.GetMSGraphAllowedProperties(def)...)
		}
		editRange := ilsp.HCLRangeToLSP(hcl.Range{Start: pos, End: pos, Filename: filename})
		candidateList = keyCandidates(keys, editRange, lastHclNode)

		if len(lastHclNode.Children) == 0 {
			propertySets := make([]schema.PropertySet, 0)
			for _, def := range defs {
				propertySets = append(propertySets, schema.GetMSGraphRequiredPropertySet(def)...)
			}
			candidateList = append(candidateList, requiredPropertiesCandidates(propertySets, editRange, lastHclNode)...)
		}
	}
	return candidateList
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
