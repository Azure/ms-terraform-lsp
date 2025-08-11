package processor

import (
	"github.com/Azure/ms-terraform-lsp/module-schema/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestConvertAttrDefaultToCtyValue tests the ConvertAttrDefaultToCtyValue function
func TestConvertAttrDefaultToCtyValue(t *testing.T) {
	t.Run("Valid cty.Value", func(t *testing.T) {
		expectedVal := cty.StringVal("hello")
		var input any = expectedVal
		result := ConvertAttrDefaultToCtyValue(input)
		if result == nil {
			t.Fatal("Expected a cty.Value, but got nil")
		}
		if !result.RawEquals(expectedVal) {
			t.Errorf("Expected %v, got %v", expectedVal, *result)
		}
	})

	t.Run("Nil input", func(t *testing.T) {
		if ConvertAttrDefaultToCtyValue(nil) != nil {
			t.Error("Expected nil for nil input")
		}
	})

	t.Run("String input", func(t *testing.T) {
		input := "this is a string"
		expected := cty.StringVal(input)
		result := ConvertAttrDefaultToCtyValue(input)
		if result == nil {
			t.Fatal("Expected a cty.Value, but got nil")
		}
		if !result.RawEquals(expected) {
			t.Errorf("Expected %v, got %v", expected, *result)
		}
	})
}

// loadVariablesFromFile parses a .hcl file and returns a map of variable schemas.
func loadVariablesFromFile(t *testing.T, path string) (map[string]*schema.SchemaAttribute, hcl.Diagnostics) {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %s", path, err)
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(src, filepath.Base(path))
	if diags.HasErrors() {
		return nil, diags
	}

	attributes, err := schema.FromHCLBody(file.Body)
	if err != nil {
		t.Fatalf("failed to build variables from file: %v", err)
	}
	return attributes, nil
}

// TestFindFieldsFromBlock tests the main function using data from variables.hcl
func TestFindFieldsFromBlock(t *testing.T) {
	// Load the variables from the actual .hcl file
	variableAttrs, diags := loadVariablesFromFile(t, "../variables.hcl")
	if diags.HasErrors() {
		t.Fatalf("Failed to load variables: %s", diags.Error())
	}

	// Select the 'core_config' variable attribute for our test
	coreConfigAttr, ok := variableAttrs["core_config"]
	if !ok {
		t.Fatal("Variable 'core_config' not found in variables.hcl")
	}

	// Create the input SchemaBlock for FindFieldsFromBlock
	inputBlock := &schema.SchemaBlock{
		Attributes: map[string]*schema.SchemaAttribute{
			"core_config": coreConfigAttr,
		},
	}

	// Run the function to test
	resultBlock, err := FindFieldsFromBlock(inputBlock, "test_module")
	if err != nil {
		t.Fatalf("FindFieldsFromBlock() returned an unexpected error: %v", err)
	}

	if resultBlock == nil {
		t.Fatal("FindFieldsFromBlock() returned a nil block")
	}

	// --- Assertions ---
	processedAttr := resultBlock.Attributes["core_config"]
	if processedAttr == nil {
		t.Fatal("Expected to find 'core_config' attribute in the result")
	}

	// Check AttributePath
	if processedAttr.AttributePath != "core_config" {
		t.Errorf("Incorrect AttributePath for 'core_config'. got=%s, want=%s", processedAttr.AttributePath, "core_config")
	}

	// Check top-level fields
	if _, ok := processedAttr.Fields["acr"]; !ok {
		t.Error("Expected field 'acr' not found")
	}
	if _, ok := processedAttr.Fields["storage"]; !ok {
		t.Error("Expected field 'storage' not found")
	}
	if _, ok := processedAttr.Fields["key_vault"]; !ok {
		t.Error("Expected field 'key_vault' not found")
	}
	if _, ok := processedAttr.Fields["ai_hub"]; !ok {
		t.Error("Expected field 'ai_hub' not found")
	}

	// Check nested fields and their default values
	aiHubFields := processedAttr.Fields["ai_hub"].Fields
	if aiHubFields == nil {
		t.Fatal("Fields for 'ai_hub' are nil")
	}

	// Check ai_hub.description
	descriptionAttr := aiHubFields["description"]
	if descriptionAttr == nil {
		t.Fatal("Field 'ai_hub.description' not found")
	}
	expectedDesc := "AI Hub"
	if !reflect.DeepEqual(descriptionAttr.Default, expectedDesc) {
		t.Errorf("Incorrect default for 'ai_hub.description'. got=%v, want=%v", descriptionAttr.Default, expectedDesc)
	}
	if descriptionAttr.AttributePath != "core_config.ai_hub.description" {
		t.Errorf("Incorrect AttributePath for 'ai_hub.description'. got=%s, want=%s", descriptionAttr.AttributePath, "core_config.ai_hub.description")
	}

	// Check ai_hub.deploy_private_dns
	deployPrivateDnsAttr := aiHubFields["deploy_private_dns"]
	if deployPrivateDnsAttr == nil {
		t.Fatal("Field 'ai_hub.deploy_private_dns' not found")
	}
	if !reflect.DeepEqual(deployPrivateDnsAttr.Default, true) {
		t.Errorf("Incorrect default for 'ai_hub.deploy_private_dns'. got=%v, want=%v", deployPrivateDnsAttr.Default, true)
	}
	if deployPrivateDnsAttr.AttributePath != "core_config.ai_hub.deploy_private_dns" {
		t.Errorf("Incorrect AttributePath for 'ai_hub.deploy_private_dns'. got=%s, want=%s", deployPrivateDnsAttr.AttributePath, "core_config.ai_hub.deploy_private_dns")
	}
}

func TestBuildFieldsRecursivelyWithComplexObject(t *testing.T) {
	complexObjectType := cty.Object(map[string]cty.Type{
		"name": cty.String,
		"config": cty.Object(map[string]cty.Type{
			"enabled": cty.Bool,
			"ports":   cty.List(cty.Number),
		}),
		"tags": cty.Map(cty.String),
	})

	defaultValue := cty.ObjectVal(map[string]cty.Value{
		"name": cty.StringVal("test-server"),
		"config": cty.ObjectVal(map[string]cty.Value{
			"enabled": cty.True,
			"ports":   cty.ListVal([]cty.Value{cty.NumberIntVal(80), cty.NumberIntVal(443)}),
		}),
		"tags": cty.MapVal(map[string]cty.Value{
			"env":  cty.StringVal("production"),
			"team": cty.StringVal("backend"),
		}),
	})

	fields, err := buildFieldsRecursively(complexObjectType, &defaultValue, "root", "test_module")
	if err != nil {
		t.Fatalf("buildFieldsRecursively() failed: %v", err)
	}

	if fields == nil {
		t.Fatal("Expected fields map, got nil")
	}

	// Check name
	if nameAttr, ok := fields["name"]; !ok {
		t.Error("Expected 'name' field, but not found")
	} else if nameAttr.Default.(string) != "test-server" {
		t.Errorf("Incorrect default for 'name': got %v, want 'test-server'", nameAttr.Default)
	} else if nameAttr.AttributePath != "root.name" {
		t.Errorf("Incorrect AttributePath for 'name': got %s, want 'root.name'", nameAttr.AttributePath)
	}

	// Check config.enabled
	if configAttr, ok := fields["config"]; !ok {
		t.Error("Expected 'config' field, but not found")
	} else {
		if configAttr.AttributePath != "root.config" {
			t.Errorf("Incorrect AttributePath for 'config': got %s, want 'root.config'", configAttr.AttributePath)
		}
		if enabledAttr, ok := configAttr.Fields["enabled"]; !ok {
			t.Error("Expected 'enabled' field in config, but not found")
		} else if enabledAttr.Default.(bool) != true {
			t.Errorf("Incorrect default for 'enabled': got %v, want true", enabledAttr.Default)
		} else if enabledAttr.AttributePath != "root.config.enabled" {
			t.Errorf("Incorrect AttributePath for 'enabled': got %s, want 'root.config.enabled'", enabledAttr.AttributePath)
		}
	}
}

func TestBuildFieldsRecursivelyWithOptionalAttributes(t *testing.T) {
	objectWithOptional := cty.ObjectWithOptionalAttrs(
		map[string]cty.Type{
			"required_field": cty.String,
			"optional_field": cty.Number,
		},
		[]string{"optional_field"},
	)

	defaultValue := cty.ObjectVal(map[string]cty.Value{
		"required_field": cty.StringVal("i-am-required"),
	})

	fields, err := buildFieldsRecursively(objectWithOptional, &defaultValue, "root", "test_module")
	if err != nil {
		t.Fatalf("buildFieldsRecursively() failed: %v", err)
	}

	if required, ok := fields["required_field"]; !ok {
		t.Error("Missing required_field")
	} else if !required.Required {
		t.Error("required_field should be marked as required")
	} else if required.AttributePath != "root.required_field" {
		t.Errorf("Incorrect AttributePath for 'required_field': got %s, want 'root.required_field'", required.AttributePath)
	}

	if optional, ok := fields["optional_field"]; !ok {
		t.Error("Missing optional_field")
	} else if !optional.Optional {
		t.Error("optional_field should be marked as optional")
	} else if optional.AttributePath != "root.optional_field" {
		t.Errorf("Incorrect AttributePath for 'optional_field': got %s, want 'root.optional_field'", optional.AttributePath)
	}
}
