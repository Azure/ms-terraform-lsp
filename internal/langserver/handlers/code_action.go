package handlers

import (
	"context"
	"encoding/json"
	"strings"

	lsctx "github.com/Azure/ms-terraform-lsp/internal/context"
	ilsp "github.com/Azure/ms-terraform-lsp/internal/lsp"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (h *logHandler) TextDocumentCodeAction(ctx context.Context, params lsp.CodeActionParams) []lsp.CodeAction {
	ca, err := h.textDocumentCodeAction(ctx, params)
	if err != nil {
		h.logger.Printf("code action failed: %s", err)
	}

	return ca
}

func (h *logHandler) textDocumentCodeAction(ctx context.Context, params lsp.CodeActionParams) ([]lsp.CodeAction, error) {
	var list []lsp.CodeAction

	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return list, err
	}

	doc, err := fs.GetDocument(ilsp.FileHandlerFromDocumentURI(params.TextDocument.URI))
	if err != nil {
		return list, err
	}

	startDocPos := lsp.TextDocumentPositionParams{
		TextDocument: params.TextDocument,
		Position:     params.Range.Start,
	}
	startPos, err := ilsp.FilePositionFromDocumentPosition(startDocPos, doc)
	if err != nil {
		return list, err
	}

	endDocPos := lsp.TextDocumentPositionParams{
		TextDocument: params.TextDocument,
		Position:     params.Range.End,
	}
	endPos, err := ilsp.FilePositionFromDocumentPosition(endDocPos, doc)
	if err != nil {
		return list, err
	}

	data, err := doc.Text()
	if err != nil {
		return list, err
	}

	hclDoc, _ := hclsyntax.ParseConfig(data, "", hcl.InitialPos)

	body, isHcl := hclDoc.Body.(*hclsyntax.Body)
	if !isHcl {
		h.logger.Printf("file is not hcl")
		return list, nil
	}

	hasAzapiForGeneratingPermission := false
	hasAzurermForGeneratingPermission := false
	hasAzapiResources := false
	hasAzurermResources := false
	for _, block := range body.Blocks {
		if startPos.Position().Byte <= block.Range().Start.Byte && block.Range().End.Byte <= endPos.Position().Byte {
			address := strings.Join(block.Labels, ".")

			if strings.HasPrefix(address, "azurerm") {
				hasAzurermForGeneratingPermission = true
			}
			if strings.HasPrefix(address, "azapi_resource") || strings.HasPrefix(address, "azapi_update_resource") {
				// azapi_resource (resource/datasource), azapi_resource_action, azapi_update_resource
				hasAzapiForGeneratingPermission = true
			}

			if block.Type != "resource" {
				continue
			}

			if strings.HasPrefix(address, "azapi_resource.") {
				hasAzapiResources = true
			}
			if strings.HasPrefix(address, "azurerm") {
				hasAzurermResources = true
			}
		}
	}

	list = append(list, listCodeActionForGeneratingPermission(params, hasAzapiForGeneratingPermission, hasAzurermForGeneratingPermission)...)
	list = append(list, listCodeActionForMigratingResources(params, hasAzapiResources, hasAzurermResources)...)
	return list, nil
}

func listCodeActionForGeneratingPermission(params lsp.CodeActionParams, hasAzapi bool, hasAzurerm bool) []lsp.CodeAction {
	if !hasAzapi && !hasAzurerm {
		return nil
	}
	argument, _ := json.Marshal(params)
	forAllSetting, _ := json.Marshal(map[string]interface{}{
		"generateForMissingPermission": false,
	})

	forMissingSetting, _ := json.Marshal(map[string]interface{}{
		"generateForMissingPermission": true,
	})
	return []lsp.CodeAction{
		{
			Title:       "Generate Custom Role",
			Kind:        "refactor.rewrite",
			Diagnostics: nil,
			IsPreferred: false,
			Disabled:    nil,
			Edit: lsp.WorkspaceEdit{
				Changes:           nil,
				DocumentChanges:   nil,
				ChangeAnnotations: nil,
			},
			Command: &lsp.Command{
				Title:   "Generate Custom Role",
				Command: CommandAztfAuthorize,
				Arguments: []json.RawMessage{
					argument,
					forAllSetting,
				},
			},
			Data: nil,
		},
		{
			Title:       "Generate Custom Role for Missing Permissions",
			Kind:        "refactor.rewrite",
			Diagnostics: nil,
			IsPreferred: false,
			Disabled:    nil,
			Edit: lsp.WorkspaceEdit{
				Changes:           nil,
				DocumentChanges:   nil,
				ChangeAnnotations: nil,
			},
			Command: &lsp.Command{
				Title:   "Generate Custom Role for Missing Permissions",
				Command: CommandAztfAuthorize,
				Arguments: []json.RawMessage{
					argument,
					forMissingSetting,
				},
			},
			Data: nil,
		},
	}
}

func listCodeActionForMigratingResources(params lsp.CodeActionParams, hasAzapiResources bool, hasAzurermResources bool) []lsp.CodeAction {
	// If the file has both azapi and azurerm resources or neither, we can't migrate
	if hasAzapiResources == hasAzurermResources {
		return nil
	}

	title := ""
	if hasAzapiResources {
		title = "Migrate to AzureRM Provider"
	} else {
		title = "Migrate to AzAPI Provider"
	}

	argument, _ := json.Marshal(params)
	return []lsp.CodeAction{
		{
			Title:       title,
			Kind:        "refactor.rewrite",
			Diagnostics: nil,
			IsPreferred: false,
			Disabled:    nil,
			Edit: lsp.WorkspaceEdit{
				Changes:           nil,
				DocumentChanges:   nil,
				ChangeAnnotations: nil,
			},
			Command: &lsp.Command{
				Title:   title,
				Command: CommandAztfMigrate,
				Arguments: []json.RawMessage{
					argument,
				},
			},
			Data: nil,
		},
	}
}
