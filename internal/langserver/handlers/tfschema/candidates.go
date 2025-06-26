package tfschema

import (
	"fmt"
	"strings"

	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
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
