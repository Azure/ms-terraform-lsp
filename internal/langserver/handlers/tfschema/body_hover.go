package tfschema

import (
	"github.com/Azure/ms-terraform-lsp/internal/langserver/schema"
	"github.com/Azure/ms-terraform-lsp/internal/msgraph"
	"github.com/Azure/ms-terraform-lsp/internal/parser"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func msgraphBodyHover(block *hclsyntax.Block, attribute *hclsyntax.Attribute, pos hcl.Pos, data []byte) *lsp.Hover {
	urlValue := parser.ExtractMSGraphUrl(block, data)
	apiVersion := "v1.0"
	if v := parser.BlockAttributeLiteralValue(block, "api_version"); v != nil {
		apiVersion = *v
	}
	bodyDef := msgraph.SchemaLoader.GetResourceDefinition(apiVersion, urlValue)

	if bodyDef == nil {
		return nil
	}

	tokens, _ := hclsyntax.LexExpression(data[attribute.Expr.Range().Start.Byte:attribute.Expr.Range().End.Byte], "main.tf", attribute.Expr.Range().Start)
	hclNode := parser.BuildHclNode(tokens)

	if hclNode == nil {
		return nil
	}

	hclNodes := parser.HclNodeArraysOfPos(hclNode, pos)
	if len(hclNodes) == 0 {
		return nil
	}
	lastHclNode := hclNodes[len(hclNodes)-1]

	if parser.ContainsPos(lastHclNode.KeyRange, pos) {
		defs := schema.GetMSGraphDef(bodyDef.AsTypeBase(), hclNodes[0:len(hclNodes)-1], 0)
		props := make([]schema.Property, 0)
		for _, def := range defs {
			props = append(props, schema.GetMSGraphAllowedProperties(def)...)
		}
		if len(props) != 0 {
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
			p := &Property{
				Name:        props[index].Name,
				Modifier:    string(props[index].Modifier),
				Type:        props[index].Type,
				Description: props[index].Description,
			}

			return p.ToHover(lastHclNode.KeyRange)
		}
	}
	return nil
}
