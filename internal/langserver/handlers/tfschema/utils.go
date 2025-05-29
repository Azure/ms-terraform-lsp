package tfschema

import (
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
)

func GetResourceSchema(name string) *Resource {
	for _, r := range Resources {
		if r.Match(name) {
			return &r
		}
	}
	return nil
}

var _ ValueCandidatesFunc = FixedValueCandidatesFunc(nil)

func FixedValueCandidatesFunc(fixedItems []lsp.CompletionItem) ValueCandidatesFunc {
	return func(prefix *string, r lsp.Range) []lsp.CompletionItem {
		for i := range fixedItems {
			if fixedItems[i].TextEdit != nil {
				fixedItems[i].TextEdit.Range = r
			}
		}
		return fixedItems
	}
}
