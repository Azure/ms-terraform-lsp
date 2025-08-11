package schema

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// hclStringToBody is a test helper to parse an HCL string into an hcl.Body
func hclStringToBody(t *testing.T, hclString string) hcl.Body {
	t.Helper()
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(hclString), "test.hcl")
	if diags.HasErrors() {
		t.Fatalf("Failed to parse HCL: %s", diags.Error())
	}
	return file.Body
}

func TestFromHCLBody(t *testing.T) {
	hclString := `
variable "name" {
  type        = string
  description = "The name of the resource."
  default     = "default-name"
}

variable "tags" {
  type    = map(string)
  default = {
    "env" = "test"
  }
}
`
	body := hclStringToBody(t, hclString)
	attributes, err := FromHCLBody(body)
	if err != nil {
		t.Fatalf("FromHCLBody() returned an unexpected error: %v", err)
	}

	if len(attributes) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(attributes))
	}

	// Test "name" variable
	nameAttr, ok := attributes["name"]
	if !ok {
		t.Fatal("Attribute 'name' not found")
	}
	if nameAttr.Name != "name" {
		t.Errorf("Expected name 'name', got '%s'", nameAttr.Name)
	}
	if nameAttr.AttributeType != cty.String {
		t.Errorf("Expected type 'string', got '%s'", nameAttr.AttributeType.FriendlyName())
	}
	if nameAttr.Content != "The name of the resource." {
		t.Errorf("Incorrect description for 'name'")
	}
	if nameAttr.Default != "default-name" {
		t.Errorf("Incorrect default for 'name'. got=%v", nameAttr.Default)
	}

	// Test "tags" variable
	tagsAttr, ok := attributes["tags"]
	if !ok {
		t.Fatal("Attribute 'tags' not found")
	}
	// The HCL parser reads the default block as an object literal, so we expect an ObjectVal here.
	expectedDefault := cty.ObjectVal(map[string]cty.Value{"env": cty.StringVal("test")})
	if !tagsAttr.Default.(cty.Value).RawEquals(expectedDefault) {
		t.Errorf("Incorrect default for 'tags'. The expected is %v, got=%v", expectedDefault, tagsAttr.Default)
	}
}

func TestFromHCLFile(t *testing.T) {
	// Create a temporary HCL file for testing
	hclContent := `
variable "enabled" {
  type    = bool
  default = true
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.hcl")
	err := os.WriteFile(filePath, []byte(hclContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Run the function to test
	schemaBlock, err := FromHCLFile(filePath)
	if err != nil {
		t.Fatalf("FromHCLFile() returned an unexpected error: %v", err)
	}

	if schemaBlock == nil {
		t.Fatal("FromHCLFile() returned a nil block")
	}

	if len(schemaBlock.Attributes) != 1 {
		t.Fatalf("Expected 1 attribute, got %d", len(schemaBlock.Attributes))
	}

	enabledAttr, ok := schemaBlock.Attributes["enabled"]
	if !ok {
		t.Fatal("Attribute 'enabled' not found")
	}
	if enabledAttr.AttributeType != cty.Bool {
		t.Errorf("Expected type bool, got %s", enabledAttr.AttributeType.FriendlyName())
	}
	if !reflect.DeepEqual(enabledAttr.Default, true) {
		t.Errorf("Incorrect default for 'enabled'. got=%v, want=true", enabledAttr.Default)
	}
}

func TestFromHCLAttributes_ComplexObject(t *testing.T) {
	hclString := `
variable "complex_object" {
  type = object({
    name   = string
    port   = number
    labels = map(string)
  })
  description = "A complex object."
  default = {
    name   = "server"
    port   = 8080
    labels = {
      "region" = "us-west"
    }
  }
}
`
	body := hclStringToBody(t, hclString)
	content, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "variable", LabelNames: []string{"name"}}},
	})
	if diags.HasErrors() {
		t.Fatalf("Failed to get partial content: %s", diags)
	}

	block := content.Blocks[0]
	blockContent, diag := block.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "type"},
			{Name: "description"},
			{Name: "default"},
		},
	})
	if diag.HasErrors() {
		t.Fatalf("Failed to get block content: %v", diag)
	}

	attribute, err := fromHCLAttributes(blockContent.Attributes, "complex_object")
	if err != nil {
		t.Fatalf("fromHCLAttributes() failed: %v", err)
	}

	if attribute.Name != "complex_object" {
		t.Errorf("Incorrect name, got %s", attribute.Name)
	}

	expectedType := cty.Object(map[string]cty.Type{
		"name":   cty.String,
		"port":   cty.Number,
		"labels": cty.Map(cty.String),
	})

	if !attribute.AttributeType.Equals(expectedType) {
		t.Errorf("Incorrect type. got=%s, want=%s", attribute.AttributeType.FriendlyName(), expectedType.FriendlyName())
	}

	defaultValue := attribute.Default.(cty.Value)
	if !defaultValue.Type().IsObjectType() {
		t.Fatalf("Expected default value to be an object, got %s", defaultValue.Type().FriendlyName())
	}
}
