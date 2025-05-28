package parser_test

import (
	"testing"

	"github.com/Azure/azurerm-lsp/internal/parser"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func Test_AttributeAtPos(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		pos   hcl.Pos
		path  string
	}{
		{
			name: "attribute in root block",
			input: `resource "msgraph_resource" "example" {
  attribute = "value"
}`,
			pos:  hcl.Pos{Line: 2, Column: 19},
			path: "attribute",
		},
		{
			name: "attribute in nested block",
			input: `resource "msgraph_resource" "example" {
  nested_block {
	attribute = "value"
  }
}`,
			pos:  hcl.Pos{Line: 3, Column: 19},
			path: "nested_block.attribute",
		},

		{
			name: "attribute in nested block",
			input: `resource "msgraph_resource" "example" {
  nested_block {
    
  }
}`,
			pos:  hcl.Pos{Line: 3, Column: 4},
			path: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			file, _ := hclsyntax.ParseConfig([]byte(tc.input), "test.hcl", hcl.InitialPos)
			body, isHcl := file.Body.(*hclsyntax.Body)
			if !isHcl {
				t.Fatalf("file is not hcl")
			}
			var block *hclsyntax.Block
			for _, nestedBlock := range body.Blocks {
				if nestedBlock.Range().ContainsPos(tc.pos) {
					block = nestedBlock
					break
				}
			}
			if block == nil {
				t.Fatalf("block not found at position %v", tc.pos)
			}
			attr, path := parser.AttributeAtPos(block, tc.pos)
			if path != tc.path {
				t.Errorf("expected path %q, got %q", tc.path, path)
			}
			if path != "" && attr == nil {
				t.Errorf("expected attribute at path %q, got nil", tc.path)
			}
		})
	}
}

func Test_BlockAtPos(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		pos   hcl.Pos
		path  string
	}{
		{
			name: "block in root",
			input: `resource "msgraph_resource" "example" {
  attribute = "value"
  
}`,
			pos:  hcl.Pos{Line: 3, Column: 3},
			path: "",
		},
		{
			name: "nested block in root",
			input: `resource "msgraph_resource" "example" {
  nested_block {
	attribute = "value"
    
  }
}`,
			pos:  hcl.Pos{Line: 4, Column: 4},
			path: "nested_block",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			file, _ := hclsyntax.ParseConfig([]byte(tc.input), "test.hcl", hcl.InitialPos)
			body, isHcl := file.Body.(*hclsyntax.Body)
			if !isHcl {
				t.Fatalf("file is not hcl")
			}
			_, path := parser.BlockAtPos(body, tc.pos)
			if path != tc.path {
				t.Errorf("expected path %q, got %q", tc.path, path)
			}
		})
	}
}
