package tfschema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/ms-terraform-lsp/internal/azure"
	"github.com/Azure/ms-terraform-lsp/internal/azure/types"
	"github.com/Azure/ms-terraform-lsp/internal/langserver/schema"
	ilsp "github.com/Azure/ms-terraform-lsp/internal/lsp"
	"github.com/Azure/ms-terraform-lsp/internal/parser"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var _ Resource = &AzAPIResource{}

type AzAPIResource struct {
	Name       string
	Properties []Property
}

func (r *AzAPIResource) GetProperty(propertyPath string) *Property {
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

func (r *AzAPIResource) Match(name string) bool {
	if r == nil {
		return false
	}
	return r.Name == name
}

func (r *AzAPIResource) ListProperties(blockPath string) []Property {
	p := r.GetProperty(blockPath)
	if p == nil {
		return nil
	}
	return p.NestedProperties
}

func (r *AzAPIResource) ResourceDocumentation(typeValue string) string {
	azureResourceType := ""
	if parts := strings.Split(typeValue, "@"); len(parts) >= 2 {
		azureResourceType = parts[0]
	}
	return fmt.Sprintf(`'%s'   
[View Documentation](https://learn.microsoft.com/en-us/azure/templates/%s?pivots=deployment-language-terraform)`, typeValue, strings.ToLower(azureResourceType))
}

func bodyCandidates(data []byte, filename string, block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, property *Property) []lsp.CompletionItem {
	if attribute.Expr != nil {
		if _, ok := attribute.Expr.(*hclsyntax.LiteralValueExpr); ok && parser.ToLiteral(attribute.Expr) == nil {
			if property != nil {
				return property.ValueCandidatesFunc(nil, editRangeFromExprRange(attribute.Expr, pos))
			}
		}
	}

	bodyDef := BodyDefinitionFromBlock(block)
	if bodyDef == nil {
		return nil
	}

	hclNode := parser.JsonEncodeExpressionToHclNode(data, attribute.Expr)
	if hclNode == nil {
		tokens, _ := hclsyntax.LexExpression(data[attribute.Expr.Range().Start.Byte:attribute.Expr.Range().End.Byte], filename, attribute.Expr.Range().Start)
		hclNode = parser.BuildHclNode(tokens)
	}
	if hclNode == nil {
		return nil
	}

	return buildAzAPIBodyCandidates(hclNode, filename, pos, bodyDef)
}

func buildAzAPIBodyCandidates(hclNode *parser.HclNode, filename string, pos hcl.Pos, def types.TypeBase) []lsp.CompletionItem {
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
		defs := schema.GetAzAPIDef(def.AsTypeBase(), hclNodes, 0)
		keys := make([]schema.Property, 0)
		for _, def := range defs {
			keys = append(keys, schema.GetAzAPIAllowedProperties(def)...)
		}
		if len(hclNodes) == 1 {
			keys = ignorePulledOutProperties(keys)
		}
		editRange := ilsp.HCLRangeToLSP(lastHclNode.KeyRange)
		candidateList = keyCandidates(keys, editRange, lastHclNode)
	case !lastHclNode.KeyRange.Empty() && !lastHclNode.EqualRange.Empty() && lastHclNode.Children == nil:
		// input property =
		defs := schema.GetAzAPIDef(def.AsTypeBase(), hclNodes, 0)
		values := make([]string, 0)
		for _, def := range defs {
			values = append(values, schema.GetAzAPIAllowedValues(def)...)
		}
		editRange := lastHclNode.ValueRange
		if lastHclNode.Value == nil {
			editRange.End = pos
		}
		candidateList = valueCandidates(values, ilsp.HCLRangeToLSP(editRange), false)
	case parser.ContainsPos(lastHclNode.ValueRange, pos):
		// input a property
		defs := schema.GetAzAPIDef(def.AsTypeBase(), hclNodes, 0)
		keys := make([]schema.Property, 0)
		for _, def := range defs {
			keys = append(keys, schema.GetAzAPIAllowedProperties(def)...)
		}
		if len(hclNodes) == 1 {
			keys = ignorePulledOutProperties(keys)
		}
		editRange := ilsp.HCLRangeToLSP(hcl.Range{Start: pos, End: pos, Filename: filename})
		candidateList = keyCandidates(keys, editRange, lastHclNode)

		if len(lastHclNode.Children) == 0 {
			propertySets := make([]schema.PropertySet, 0)
			for _, def := range defs {
				propertySets = append(propertySets, schema.GetAzAPIRequiredPropertySet(def)...)
			}
			if len(hclNodes) == 1 {
				for i, ps := range propertySets {
					propertySets[i].Name = ""
					propertySets[i].Properties = ignorePulledOutPropertiesFromPropertySet(ps.Properties)
				}
			}
			candidateList = append(candidateList, requiredPropertiesCandidates(propertySets, editRange, lastHclNode)...)
		}
	}
	return candidateList
}

func BodyDefinitionFromBlock(block *hclsyntax.Block) types.TypeBase {
	typeValue := parser.ExtractAzureResourceType(block)
	if typeValue == nil {
		return nil
	}
	var bodyDef types.TypeBase
	def, err := azure.GetResourceDefinitionByResourceType(*typeValue)
	if err != nil || def == nil {
		return nil
	}
	bodyDef = def
	if len(block.Labels) >= 2 && block.Labels[0] == "azapi_resource_action" {
		parts := strings.Split(*typeValue, "@")
		if len(parts) != 2 {
			return nil
		}
		actionName := parser.ExtractAction(block)
		if actionName != nil && len(*actionName) != 0 {
			resourceFuncDef, err := azure.GetResourceFunction(parts[0], parts[1], *actionName)
			if err != nil || resourceFuncDef == nil {
				return nil
			}
			bodyDef = resourceFuncDef
		}
	}
	return bodyDef
}

func ignorePulledOutProperties(input []schema.Property) []schema.Property {
	res := make([]schema.Property, 0)
	// ignore properties pulled out from body
	for _, p := range input {
		if !isPropertyPulledOut(p) {
			res = append(res, p)
		}
	}
	return res
}

func ignorePulledOutPropertiesFromPropertySet(properties map[string]schema.Property) map[string]schema.Property {
	res := make(map[string]schema.Property)
	// ignore properties pulled out from body
	for _, p := range properties {
		if !isPropertyPulledOut(p) {
			res[p.Name] = p
		}
	}
	return res
}

func isPropertyPulledOut(p schema.Property) bool {
	return p.Name == "name" || p.Name == "location" || p.Name == "identity" || p.Name == "tags"
}

func typeCandidates(prefix *string, r lsp.Range) []lsp.CompletionItem {
	candidates := make([]lsp.CompletionItem, 0)
	if prefix == nil || !strings.Contains(*prefix, "@") {
		for resourceType := range azure.GetAzureSchema().Resources {
			candidates = append(candidates, lsp.CompletionItem{
				Label: fmt.Sprintf(`"%s"`, resourceType),
				Kind:  lsp.ValueCompletion,
				Documentation: lsp.MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("Type: `%s`  \n", resourceType),
				},
				SortText:         resourceType,
				InsertTextFormat: lsp.SnippetTextFormat,
				InsertTextMode:   lsp.AdjustIndentation,
				TextEdit: &lsp.TextEdit{
					Range:   r,
					NewText: fmt.Sprintf(`"%s@$0"`, resourceType),
				},
				Command: constTriggerSuggestCommand(),
			})
		}
	} else {
		resourceType := (*prefix)[0:strings.Index(*prefix, "@")]
		apiVersions := azure.GetApiVersions(resourceType)
		sort.Strings(apiVersions)
		length := len(apiVersions)
		for index, apiVersion := range apiVersions {
			candidates = append(candidates, lsp.CompletionItem{
				Label: fmt.Sprintf(`"%s@%s"`, resourceType, apiVersion),
				Kind:  lsp.ValueCompletion,
				Documentation: lsp.MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("Type: `%s`  \nAPI Version: `%s`", resourceType, apiVersion),
				},
				SortText:         fmt.Sprintf("%04d", length-index),
				InsertTextFormat: lsp.PlainTextTextFormat,
				InsertTextMode:   lsp.AdjustIndentation,
				TextEdit: &lsp.TextEdit{
					Range:   r,
					NewText: fmt.Sprintf(`"%s@%s"`, resourceType, apiVersion),
				},
			})
		}
	}

	return candidates
}

func locationCandidates(_ *string, r lsp.Range) []lsp.CompletionItem {
	values := make([]string, 0)
	for _, location := range supportedLocations() {
		values = append(values, fmt.Sprintf(`"%s"`, location))
	}
	return valueCandidates(values, r, true)
}

func identityTypesCandidates(_ *string, r lsp.Range) []lsp.CompletionItem {
	values := []string{
		`"SystemAssigned"`,
		`"UserAssigned"`,
		`"SystemAssigned, UserAssigned"`,
	}
	return valueCandidates(values, r, false)
}

func boolCandidates(_ *string, r lsp.Range) []lsp.CompletionItem {
	return valueCandidates([]string{"true", "false"}, r, false)
}

func resourceHttpMethodCandidates(_ *string, r lsp.Range) []lsp.CompletionItem {
	return valueCandidates([]string{
		`"POST"`,
		`"PATCH"`,
		`"PUT"`,
		`"DELETE"`,
	}, r, true)
}

func dataSourceHttpMethodCandidates(_ *string, r lsp.Range) []lsp.CompletionItem {
	return valueCandidates([]string{
		`"POST"`,
		`"GET"`,
	}, r, true)
}

func actionCandidates(_ []byte, _ string, block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, _ *Property) []lsp.CompletionItem {
	typeValue := parser.ExtractAzureResourceType(block)
	if typeValue == nil {
		return nil
	}

	parts := strings.Split(*typeValue, "@")
	if len(parts) != 2 {
		return nil
	}

	functions, err := azure.ListResourceFunctions(parts[0], parts[1])
	if err != nil {
		return nil
	}

	values := make([]string, 0)
	for _, function := range functions {
		def, err := function.GetDefinition()
		if err == nil && def != nil {
			values = append(values, fmt.Sprintf(`"%s"`, def.Name))
		}
	}

	return valueCandidates(values, editRangeFromExprRange(attribute.Expr, pos), false)
}

func azapiBodyHover(block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, data []byte) *lsp.Hover {
	bodyDef := BodyDefinitionFromBlock(block)
	if bodyDef == nil {
		return nil
	}

	hclNode := parser.JsonEncodeExpressionToHclNode(data, attribute.Expr)
	if hclNode == nil {
		tokens, _ := hclsyntax.LexExpression(data[attribute.Expr.Range().Start.Byte:attribute.Expr.Range().End.Byte], "main.tf", attribute.Expr.Range().Start)
		hclNode = parser.BuildHclNode(tokens)
	}
	if hclNode == nil {
		return nil
	}

	hclNodes := parser.HclNodeArraysOfPos(hclNode, pos)
	if len(hclNodes) == 0 {
		return nil
	}
	lastHclNode := hclNodes[len(hclNodes)-1]
	if !parser.ContainsPos(lastHclNode.KeyRange, pos) {
		return nil
	}

	defs := schema.GetAzAPIDef(bodyDef.AsTypeBase(), hclNodes[0:len(hclNodes)-1], 0)
	props := make([]schema.Property, 0)
	for _, def := range defs {
		props = append(props, schema.GetAzAPIAllowedProperties(def)...)
	}
	if len(props) == 0 {
		return nil
	}

	index := -1
	for i := range props {
		if props[i].Name == lastHclNode.Key {
			index = i
			break
		}
	}
	if index == -1 {
		return nil
	}

	return &lsp.Hover{
		Range: ilsp.HCLRangeToLSP(lastHclNode.KeyRange),
		Contents: lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: fmt.Sprintf("```\n%s: %s(%s)\n```\n%s", props[index].Name, string(props[index].Modifier), props[index].Type, props[index].Description),
		},
	}
}
