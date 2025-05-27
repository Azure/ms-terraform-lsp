package parser

import (
	"context"
	"fmt"
	"strings"

	lsctx "github.com/Azure/azurerm-lsp/internal/context"
	"github.com/Azure/azurerm-lsp/internal/lsp"
	"github.com/Azure/azurerm-lsp/internal/protocol"
)

func GetDocumentContent(ctx context.Context, documentURI protocol.DocumentURI) (string, string, error) {
	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return "", "", err
	}
	doc, err := fs.GetDocument(lsp.FileHandlerFromDocumentURI(documentURI))
	if err != nil {
		return "", "", fmt.Errorf("failed to get document: %v", err)
	}
	data, err := doc.Text()
	if err != nil {
		return "", "", fmt.Errorf("failed to get document text: %v", err)
	}
	return string(data), doc.Filename(), nil
}

func GetLineContent(contents []string, line uint32) (string, error) {
	if int(line) >= len(contents) {
		return "", fmt.Errorf("invalid line number")
	}
	return strings.TrimSpace(contents[line]), nil
}
