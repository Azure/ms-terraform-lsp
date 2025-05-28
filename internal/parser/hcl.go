package parser

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"regexp"
	"strings"
)

func ToLiteral(expression hclsyntax.Expression) *string {
	value, dialog := expression.Value(&hcl.EvalContext{})
	if dialog != nil && dialog.HasErrors() {
		return nil
	}
	if value.Type() == cty.String && !value.IsNull() && value.IsKnown() {
		v := value.AsString()
		return &v
	}
	return nil
}

func ToLiteralBoolean(expression hclsyntax.Expression) *bool {
	value, dialog := expression.Value(&hcl.EvalContext{})
	if dialog != nil && dialog.HasErrors() {
		return nil
	}
	if value.Type() == cty.Bool && !value.IsNull() && value.IsKnown() {
		v := value.True()
		return &v
	}
	return nil
}

func BlockAtPos(body *hclsyntax.Body, pos hcl.Pos) (*hclsyntax.Block, string) {
	for _, b := range body.Blocks {
		blockType := ""
		if b.Type != "data" && b.Type != "resource" && len(b.Labels) == 0 {
			blockType = b.Type
		}
		if ContainsPos(b.Range(), pos) {
			block, blockName := BlockAtPos(b.Body, pos)
			if block != nil {
				if blockType == "" {
					return block, blockName
				}
				return block, fmt.Sprintf("%s.%s", blockType, blockName)
			}
			return b, blockType
		}
	}
	return nil, ""
}

func AttributeAtPos(block *hclsyntax.Block, pos hcl.Pos) (*hclsyntax.Attribute, string) {
	if block == nil {
		return nil, ""
	}

	for _, attr := range block.Body.Attributes {
		if ContainsPos(attr.SrcRange, pos) {
			return attr, attr.Name
		}
		if ContainsPos(attr.Expr.Range(), pos) {
			return attr, attr.Name
		}
	}
	for _, nestedBlock := range block.Body.Blocks {
		if ContainsPos(nestedBlock.Range(), pos) {
			attr, attrName := AttributeAtPos(nestedBlock, pos)
			if attr != nil {
				return attr, fmt.Sprintf("%s.%s", nestedBlock.Type, attrName)
			}
		}
	}

	return nil, ""
}

func AttributeWithName(block *hclsyntax.Block, name string) *hclsyntax.Attribute {
	if block == nil {
		return nil
	}
	for _, attr := range block.Body.Attributes {
		if attr.Name == name {
			return attr
		}
	}
	return nil
}

func BlockAttributeLiteralValue(block *hclsyntax.Block, name string) *string {
	attr := AttributeWithName(block, name)
	if attr == nil {
		return nil
	}
	return ToLiteral(attr.Expr)
}

func ExtractMSGraphUrl(block *hclsyntax.Block, data []byte) string {
	urlValue := urlValueFromComment(data, block.Range())
	if urlValue == "" {
		if v := BlockAttributeLiteralValue(block, "url"); v != nil {
			urlValue = *v
		} else {
			return ""
		}
	}
	urlValue = strings.ReplaceAll(urlValue, "${", "{")
	return urlValue
}

var urlRegex = regexp.MustCompile(`url\s*=\s*"(.+)"`)

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
