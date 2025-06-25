package tfschema

import (
	"fmt"
	"github.com/Azure/ms-terraform-lsp/internal/langserver/schema"
	ilsp "github.com/Azure/ms-terraform-lsp/internal/lsp"
	"github.com/Azure/ms-terraform-lsp/internal/parser"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

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

func editRangeFromExprRange(expression hclsyntax.Expression, pos hcl.Pos) lsp.Range {
	expRange := expression.Range()
	if expRange.Start.Line != expRange.End.Line && expRange.End.Column == 1 && expRange.End.Line-1 == pos.Line {
		expRange.End = pos
	}
	return ilsp.HCLRangeToLSP(expRange)
}
