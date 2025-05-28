package handlers

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	lsctx "github.com/Azure/azurerm-lsp/internal/context"
	"github.com/Azure/azurerm-lsp/internal/langserver/handlers/snippets"
	"github.com/Azure/azurerm-lsp/internal/langserver/handlers/tfschema"
	ilsp "github.com/Azure/azurerm-lsp/internal/lsp"
	"github.com/Azure/azurerm-lsp/internal/parser"
	lsp "github.com/Azure/azurerm-lsp/internal/protocol"
	"github.com/Azure/azurerm-lsp/internal/utils"
	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (svc *service) TextDocumentComplete(ctx context.Context, params lsp.CompletionParams) (lsp.CompletionList, error) {
	var list lsp.CompletionList

	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return list, err
	}

	_, err = ilsp.ClientCapabilities(ctx)
	if err != nil {
		return list, err
	}

	doc, err := fs.GetDocument(ilsp.FileHandlerFromDocumentURI(params.TextDocument.URI))
	if err != nil {
		return list, err
	}

	fPos, err := ilsp.FilePositionFromDocumentPosition(params.TextDocumentPositionParams, doc)
	if err != nil {
		return list, err
	}

	svc.logger.Printf("Looking for candidates at %q -> %#v", doc.Filename(), fPos.Position())

	data, err := doc.Text()
	if err != nil {
		return list, err
	}

	candidates := CandidatesAtPos(data, doc.Filename(), fPos.Position(), svc.logger)
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].SortText < candidates[j].SortText })

	return lsp.CompletionList{
		IsIncomplete: false,
		Items:        candidates,
	}, nil
}

func CandidatesAtPos(data []byte, filename string, pos hcl.Pos, logger *log.Logger) []lsp.CompletionItem {
	file, _ := hclsyntax.ParseConfig(data, filename, hcl.InitialPos)

	body, isHcl := file.Body.(*hclsyntax.Body)
	if !isHcl {
		logger.Printf("file is not hcl")
		return nil
	}

	candidateList := make([]lsp.CompletionItem, 0)

	var resourceBlock *hclsyntax.Block
	for _, block := range body.Blocks {
		if parser.ContainsPos(block.Range(), pos) {
			resourceBlock = block
			break
		}
	}

	// the cursor is not in a block
	if resourceBlock == nil {
		editRange := lsp.Range{
			Start: ilsp.HCLPosToLSP(pos),
			End:   ilsp.HCLPosToLSP(pos),
		}
		editRange.Start.Character = 0

		// msgraph templates
		candidateList = append(candidateList, snippets.MSGraphTemplateCandidates(editRange)...)

		// azurerm templates
		if shouldGiveTopLevelCompletions(string(data), pos.Line-1) {
			candidateList = append(candidateList, snippets.AzureRMTemplateCandidates(editRange)...)
		}
		return candidateList
	}

	// if the block has no labels, we cannot provide any candidates
	if len(resourceBlock.Labels) == 0 {
		return candidateList
	}

	resourceName := fmt.Sprintf("%s.%s", resourceBlock.Type, resourceBlock.Labels[0])
	resource := tfschema.GetResourceSchema(resourceName)
	if resource == nil {
		return candidateList
	}

	// if the cursor is in an attribute, provide value candidates for that attribute
	if attribute, attributePath := parser.AttributeAtPos(resourceBlock, pos); attribute != nil {
		propertyPath := fmt.Sprintf("%s.%s", resourceName, attributePath)
		property := (*resource).GetProperty(propertyPath)
		if property == nil {
			return candidateList
		}
		if property.GenericCandidatesFunc != nil {
			candidateList = append(candidateList, property.GenericCandidatesFunc(data, filename, resourceBlock, attribute, pos, property)...)
		} else if property.ValueCandidatesFunc != nil {
			prefix := parser.ToLiteral(attribute.Expr)
			candidateList = append(candidateList, property.ValueCandidatesFunc(prefix, editRangeFromExprRange(attribute.Expr, pos))...)
		}

		return candidateList
	}

	if nestedBlock, blockPath := parser.BlockAtPos(body, pos); nestedBlock != nil {
		var editRange *lsp.Range

		if blockPath == "" {
			editRange = &lsp.Range{
				Start: ilsp.HCLPosToLSP(pos),
				End:   ilsp.HCLPosToLSP(pos),
			}
			editRange.Start.Character = 2
		}

		if blockPath == "" {
			candidateList = append(candidateList, snippets.MSGraphCodeSampleCandidates(resourceBlock, *editRange, data)...)
		}

		blockPath = fmt.Sprintf("%s.%s", resourceName, blockPath)
		candidateList = append(candidateList, tfschema.PropertiesCandidates((*resource).ListProperties(blockPath), editRange)...)
	}

	return candidateList
}

func editRangeFromExprRange(expression hclsyntax.Expression, pos hcl.Pos) lsp.Range {
	expRange := expression.Range()
	if expRange.Start.Line != expRange.End.Line && expRange.End.Column == 1 && expRange.End.Line-1 == pos.Line {
		expRange.End = pos
	}
	return ilsp.HCLRangeToLSP(expRange)
}

func shouldGiveTopLevelCompletions(content string, line int) bool {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return false
	}

	currentLine := strings.TrimSpace(lines[line])
	return utils.MatchAnyPrefix(currentLine, schema.AzureRMPrefix)
}
