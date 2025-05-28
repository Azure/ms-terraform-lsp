package tfschema

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/microsoft/msgraph-lsp/internal/langserver/schema"
	ilsp "github.com/microsoft/msgraph-lsp/internal/lsp"
	"github.com/microsoft/msgraph-lsp/internal/msgraph"
	"github.com/microsoft/msgraph-lsp/internal/parser"
	lsp "github.com/microsoft/msgraph-lsp/internal/protocol"
	"github.com/ms-henglu/go-msgraph-types/types"
)

func bodyCandidates(data []byte, filename string, block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, property *Property) []lsp.CompletionItem {
	if attribute.Expr != nil {
		if _, ok := attribute.Expr.(*hclsyntax.LiteralValueExpr); ok && parser.ToLiteral(attribute.Expr) == nil {
			if property != nil {
				return property.ValueCandidatesFunc(nil, editRangeFromExprRange(attribute.Expr, pos))
			}
		}
	}

	bodyDef := BodyDefinitionFromBlock(block, data)
	if bodyDef == nil {
		return nil
	}

	tokens, _ := hclsyntax.LexExpression(data[attribute.Expr.Range().Start.Byte:attribute.Expr.Range().End.Byte], filename, attribute.Expr.Range().Start)
	hclNode := parser.BuildHclNode(tokens)
	if hclNode == nil {
		return nil
	}

	return buildCandidates(hclNode, filename, pos, bodyDef)
}

func buildCandidates(hclNode *parser.HclNode, filename string, pos hcl.Pos, def types.TypeBase) []lsp.CompletionItem {
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
		defs := schema.GetDef(def.AsTypeBase(), hclNodes, 0)
		keys := make([]schema.Property, 0)
		for _, def := range defs {
			keys = append(keys, schema.GetAllowedProperties(def)...)
		}
		editRange := ilsp.HCLRangeToLSP(lastHclNode.KeyRange)
		candidateList = keyCandidates(keys, editRange, lastHclNode)
	case !lastHclNode.KeyRange.Empty() && !lastHclNode.EqualRange.Empty() && lastHclNode.Children == nil:
		// input property =
		defs := schema.GetDef(def.AsTypeBase(), hclNodes, 0)
		values := make([]string, 0)
		for _, def := range defs {
			values = append(values, schema.GetAllowedValues(def)...)
		}
		editRange := lastHclNode.ValueRange
		if lastHclNode.Value == nil {
			editRange.End = pos
		}
		candidateList = valueCandidates(values, ilsp.HCLRangeToLSP(editRange), false)
	case parser.ContainsPos(lastHclNode.ValueRange, pos):
		// input a property
		defs := schema.GetDef(def.AsTypeBase(), hclNodes, 0)
		keys := make([]schema.Property, 0)
		for _, def := range defs {
			keys = append(keys, schema.GetAllowedProperties(def)...)
		}
		editRange := ilsp.HCLRangeToLSP(hcl.Range{Start: pos, End: pos, Filename: filename})
		candidateList = keyCandidates(keys, editRange, lastHclNode)

		if len(lastHclNode.Children) == 0 {
			propertySets := make([]schema.PropertySet, 0)
			for _, def := range defs {
				propertySets = append(propertySets, schema.GetRequiredPropertySet(def)...)
			}
			candidateList = append(candidateList, requiredPropertiesCandidates(propertySets, editRange, lastHclNode)...)
		}
	}
	return candidateList
}

func keyCandidates(props []schema.Property, r lsp.Range, parentNode *parser.HclNode) []lsp.CompletionItem {
	candidates := make([]lsp.CompletionItem, 0)
	propSet := make(map[string]bool)
	for _, prop := range props {
		// skip the @odata.type property
		if prop.Name == "@odata.type" {
			continue
		}
		if propSet[prop.Name] {
			continue
		}
		propSet[prop.Name] = true
		content := prop.Name
		newText := ""
		sortText := fmt.Sprintf("1%s", content)
		if prop.Modifier == schema.Required {
			sortText = fmt.Sprintf("0%s", content)
		}

		keyPart := fmt.Sprintf(`%s =`, content)
		if parentNode.KeyValueFormat == parser.QuotedKeyEqualValue {
			keyPart = fmt.Sprintf(`"%s" =`, content)
		} else if parentNode.KeyValueFormat == parser.QuotedKeyColonValue {
			keyPart = fmt.Sprintf(`"%s":`, content)
		}

		switch prop.Type {
		case "string":
			newText = fmt.Sprintf(`%s "$0"`, keyPart)
		case "array":
			newText = fmt.Sprintf(`%s [$0]`, keyPart)
		case "object":
			newText = fmt.Sprintf("%s {\n\t$0\n}", keyPart)
		default:
			newText = fmt.Sprintf(`%s $0`, keyPart)
		}
		candidates = append(candidates, lsp.CompletionItem{
			Label:  content,
			Kind:   lsp.PropertyCompletion,
			Detail: fmt.Sprintf("%s (%s)", prop.Name, prop.Modifier),
			Documentation: lsp.MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("Type: `%s`  \n%s\n", prop.Type, prop.Description),
			},
			SortText:         sortText,
			InsertTextFormat: lsp.SnippetTextFormat,
			InsertTextMode:   lsp.AdjustIndentation,
			TextEdit: &lsp.TextEdit{
				Range:   r,
				NewText: newText,
			},
			Command: constTriggerSuggestCommand(),
		})
	}
	return candidates
}

func requiredPropertiesCandidates(propertySets []schema.PropertySet, r lsp.Range, parentNode *parser.HclNode) []lsp.CompletionItem {
	candidates := make([]lsp.CompletionItem, 0)
	for _, ps := range propertySets {
		if len(ps.Properties) == 0 {
			continue
		}
		props := make([]schema.Property, 0)
		for _, prop := range ps.Properties {
			props = append(props, prop)
		}
		for range props {
			for i := 0; i < len(props)-1; i++ {
				if props[i].Name > props[i+1].Name {
					props[i], props[i+1] = props[i+1], props[i]
				}
			}
		}
		newText := ""
		index := 1
		for _, prop := range props {
			keyPart := fmt.Sprintf(`%s =`, prop.Name)
			if parentNode.KeyValueFormat == parser.QuotedKeyEqualValue {
				keyPart = fmt.Sprintf(`"%s" =`, prop.Name)
			} else if parentNode.KeyValueFormat == parser.QuotedKeyColonValue {
				keyPart = fmt.Sprintf(`"%s":`, prop.Name)
			}

			if len(prop.Value) != 0 {
				newText += fmt.Sprintf("%s \"%s\"\n", keyPart, prop.Value)
			} else {
				switch prop.Type {
				case "string":
					newText += fmt.Sprintf(`%s "$%d"`, keyPart, index)
				case "array":
					newText += fmt.Sprintf(`%s [$%d]`, keyPart, index)
				case "object":
					newText += fmt.Sprintf("%s {\n\t$%d\n}", keyPart, index)
				default:
					newText += fmt.Sprintf(`%s $%d`, keyPart, index)
				}
				newText += "\n"
				index++
			}
		}

		label := "required-properties"
		if len(ps.Name) != 0 {
			label = fmt.Sprintf("required-properties-%s", ps.Name)
		}
		detail := "Required properties"
		if len(ps.Name) != 0 {
			detail = fmt.Sprintf("Required properties - %s", ps.Name)
		}
		candidates = append(candidates, lsp.CompletionItem{
			Label:  label,
			Kind:   lsp.SnippetCompletion,
			Detail: detail,
			Documentation: lsp.MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("Type: `%s`  \n```\n%s\n```\n", ps.Name, newText),
			},
			SortText:         "0",
			InsertTextFormat: lsp.SnippetTextFormat,
			InsertTextMode:   lsp.AdjustIndentation,
			TextEdit: &lsp.TextEdit{
				Range:   r,
				NewText: newText,
			},
			Command: constTriggerSuggestCommand(),
		})
	}
	return candidates
}

var urlRegex = regexp.MustCompile(`url\s*=\s*"(.+)"`)

func BodyDefinitionFromBlock(block *hclsyntax.Block, data []byte) *types.ResourceType {
	urlValue := urlValueFromComment(data, block.Range())
	if urlValue == "" {
		if v := parser.BlockAttributeLiteralValue(block, "url"); v != nil {
			urlValue = *v
		} else {
			return nil
		}
	}

	urlValue = strings.ReplaceAll(urlValue, "${", "{")

	apiVersion := "v1.0"
	if v := parser.BlockAttributeLiteralValue(block, "api_version"); v != nil {
		apiVersion = *v
	}

	return msgraph.SchemaLoader.GetResourceDefinition(apiVersion, urlValue)
}

func urlValueFromComment(data []byte, r hcl.Range) string {
	if r.End.Byte > len(data) {
		return ""
	}
	src := data[r.Start.Byte:r.End.Byte]
	matches := urlRegex.FindSubmatch(src)
	if len(matches) != 2 {
		return ""
	}
	return string(matches[1])
}

func editRangeFromExprRange(expression hclsyntax.Expression, pos hcl.Pos) lsp.Range {
	expRange := expression.Range()
	if expRange.Start.Line != expRange.End.Line && expRange.End.Column == 1 && expRange.End.Line-1 == pos.Line {
		expRange.End = pos
	}
	return ilsp.HCLRangeToLSP(expRange)
}
