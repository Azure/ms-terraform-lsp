package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	lsctx "github.com/Azure/ms-terraform-lsp/internal/context"
	"github.com/Azure/ms-terraform-lsp/internal/langserver/handlers/tfschema"
	ilsp "github.com/Azure/ms-terraform-lsp/internal/lsp"
	"github.com/Azure/ms-terraform-lsp/internal/parser"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/Azure/ms-terraform-lsp/internal/telemetry"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (svc *service) TextDocumentHover(ctx context.Context, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return nil, err
	}

	_, err = ilsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	doc, err := fs.GetDocument(ilsp.FileHandlerFromDocumentURI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}

	fPos, err := ilsp.FilePositionFromDocumentPosition(params, doc)
	if err != nil {
		return nil, err
	}

	data, err := doc.Text()
	if err != nil {
		return nil, err
	}

	telemetrySender, err := lsctx.Telemetry(ctx)
	if err != nil {
		return nil, err
	}

	svc.logger.Printf("Looking for hover data at %q -> %#v", doc.Filename(), fPos.Position())
	hoverData := HoverAtPos(ctx, data, doc.Filename(), fPos.Position(), svc.logger, telemetrySender)
	svc.logger.Printf("received hover data: %#v", hoverData)

	return hoverData, nil
}

func HoverAtPos(ctx context.Context, data []byte, filename string, pos hcl.Pos, logger *log.Logger, sender telemetry.Sender) *lsp.Hover {
	file, _ := hclsyntax.ParseConfig(data, filename, hcl.InitialPos)
	body, isHcl := file.Body.(*hclsyntax.Body)
	if !isHcl {
		logger.Printf("file is not hcl")
		return nil
	}

	var resourceBlock *hclsyntax.Block
	for _, block := range body.Blocks {
		if parser.ContainsPos(block.Range(), pos) {
			resourceBlock = block
			break
		}
	}

	// the cursor is not in a block
	if resourceBlock == nil {
		return nil
	}

	// if the block has no labels, we cannot provide any hover information
	if len(resourceBlock.Labels) == 0 {
		return nil
	}

	resourceName := fmt.Sprintf("%s.%s", resourceBlock.Type, resourceBlock.Labels[0])
	resource := tfschema.GetResourceSchema(resourceName)
	if resource == nil {
		return nil
	}

	// if the cursor is in an attribute, provide hover information for the attribute
	if attribute, attributePath := parser.AttributeAtPos(resourceBlock, pos); attribute != nil {
		propertyPath := fmt.Sprintf("%s.%s", resourceName, attributePath)
		property := (*resource).GetProperty(propertyPath)
		if property == nil {
			return nil
		}
		if property.CustomizedHoverFunc != nil {
			return property.CustomizedHoverFunc(resourceBlock, attribute, pos, data)
		}
		if !parser.ContainsPos(attribute.NameRange, pos) {
			return nil
		}
		return property.ToHover(attribute.NameRange)
	}

	// if the cursor is in a nested block, provide hover information for the block
	if nestedBlock, blockPath := parser.BlockAtPos(body, pos); nestedBlock != nil {
		if !parser.ContainsPos(nestedBlock.DefRange(), pos) {
			return nil
		}

		if blockPath != "" {
			blockPath = fmt.Sprintf("%s.%s", resourceName, blockPath)
			property := (*resource).GetProperty(blockPath)
			if property == nil {
				return nil
			}
			return property.ToHover(nestedBlock.DefRange())
		}

		// hover on the block itself
		var doc, docId string
		switch {
		case strings.Contains(resourceName, "azapi_"):
			typeValue := ""
			if v := parser.BlockAttributeLiteralValue(resourceBlock, "type"); v != nil {
				typeValue = *v
			}
			doc = (*resource).ResourceDocumentation(typeValue)
			docId = fmt.Sprintf("azapi_resource.%s", typeValue)

		case strings.Contains(resourceName, "msgraph_resource"):
			url := parser.ExtractMSGraphUrl(resourceBlock, data)
			apiVersion := "v1.0"
			if v := parser.BlockAttributeLiteralValue(resourceBlock, "api_version"); v != nil {
				apiVersion = *v
			}
			resourceType := fmt.Sprintf("%s@%s", url, apiVersion)
			doc = (*resource).ResourceDocumentation(resourceType)
			docId = fmt.Sprintf("msgraph_resource.%s", resourceType)

		case strings.Contains(resourceName, "azurerm_"):
			doc = (*resource).ResourceDocumentation(resourceName)
			docId = resourceName
		}

		if doc == "" {
			return nil
		}
		sender.SendEvent(ctx, "textDocument/hover", map[string]interface{}{
			"kind": "resource-definition",
			"url":  docId,
		})

		return &lsp.Hover{
			Range: ilsp.HCLRangeToLSP(resourceBlock.DefRange()),
			Contents: lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: doc,
			},
		}
	}

	return nil
}
