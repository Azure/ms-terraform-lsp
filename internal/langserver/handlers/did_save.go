package handlers

import (
	"context"

	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
)

func (lh *logHandler) TextDocumentDidSave(ctx context.Context, params lsp.DidSaveTextDocumentParams) error {
	return nil
}
