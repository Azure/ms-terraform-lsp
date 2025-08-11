package processor

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/ms-terraform-lsp/module-schema/schema"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessBatchOutput(t *testing.T) {
	// Create a temporary test directory structure
	tempDir := t.TempDir()

	// Create subdirectories
	combinedVarDir := filepath.Join(tempDir, "combinedVar")
	examplesDir := filepath.Join(tempDir, "examples")
	readmesDir := filepath.Join(tempDir, "readmes")

	err := os.MkdirAll(combinedVarDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(examplesDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(readmesDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test data files
	testCases := []struct {
		name        string
		objName     string
		variableHCL string
		exampleHCL  string
		readmeMD    string
	}{
		{
			name:    "test-module-1",
			objName: "test-module-1",
			variableHCL: `variable "location" {
  type        = string
  description = "The Azure region where the resource will be deployed."
  nullable    = false
}

variable "resource_group_name" {
  type        = string
  description = "The name of the resource group in which to create the resources."
  nullable    = false
}

variable "enable_telemetry" {
  type        = bool
  default     = true
  description = <<DESCRIPTION
This variable controls whether or not telemetry is enabled for the module.
For more information see <https://aka.ms/avm/telemetryinfo>.
If it is set to false, then no telemetry will be collected.
DESCRIPTION
  nullable    = false
}`,
			exampleHCL: `terraform {
  required_version = ">= 1.9"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.71, < 5.0"
    }
  }
}

provider "azurerm" {
  features {}
}

module "test_module" {
  source = "../../"
  
  location            = "East US"
  resource_group_name = "test-rg"
  enable_telemetry    = true
}`,
			readmeMD: `# Test Module

This is a test module for demonstrating functionality.

## Features

- Creates test resources
- Supports telemetry
- Configurable location

## Usage

See the examples directory for usage examples.`,
		},
		{
			name:    "test-module-2",
			objName: "test-module-2",
			variableHCL: `variable "name" {
  type        = string
  description = "The name of the resource."
  nullable    = false
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "A mapping of tags to assign to the resource."
  nullable    = false
}`,
			exampleHCL: `terraform {
  required_version = ">= 1.9"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.71, < 5.0"
    }
  }
}

provider "azurerm" {
  features {}
}

module "test_module_2" {
  source="../../"
  
  name = "test-resource"
  tags = {
    Environment = "Test"
    Purpose     = "Testing"
  }
}`,
			readmeMD: `# Test Module 2

Another test module with different configuration options.

## Configuration

This module supports:
- Custom naming
- Tag management
`,
		},
	}

	// Create test files
	for _, tc := range testCases {
		// Create variables file
		variablesFilePath := filepath.Join(combinedVarDir, fmt.Sprintf("%s_variables.tf", tc.objName))
		err := os.WriteFile(variablesFilePath, []byte(tc.variableHCL), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Create example file
		exampleFilePath := filepath.Join(examplesDir, fmt.Sprintf("%s_example.tf", tc.objName))
		err = os.WriteFile(exampleFilePath, []byte(tc.exampleHCL), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Create README file
		readmeFilePath := filepath.Join(readmesDir, fmt.Sprintf("%s_README.md", tc.objName))
		err = os.WriteFile(readmeFilePath, []byte(tc.readmeMD), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test ProcessBatchOutput
	result, err := ProcessBatchOutput(tempDir)
	if err != nil {
		t.Fatalf("ProcessBatchOutput failed: %v", err)
	}

	// Verify results
	if len(result) != len(testCases) {
		t.Errorf("Expected %d results, got %d", len(testCases), len(result))
	}

	// Check each result
	for i, tc := range testCases {
		if i >= len(result) {
			t.Errorf("Missing result for test case %s", tc.name)
			continue
		}

		obj := result[i]
		expectedName := fmt.Sprintf("Azure/%s/azurerm", tc.objName)

		// Check name
		if obj.Name != expectedName {
			t.Errorf("Expected name %s, got %s", expectedName, obj.Name)
		}

		// Check that fields exist
		if len(obj.Fields) == 0 {
			t.Errorf("Expected fields for %s, got none", tc.name)
		}

		// Check that example HCL is processed correctly (source replacement)
		if !strings.Contains(obj.ExampleHCL, fmt.Sprintf(`source = "Azure/%s/azurerm"`, tc.objName)) {
			t.Errorf("Expected source replacement in example HCL for %s", tc.name)
		}

		// Check that README content is included
		if obj.Details == "" {
			t.Errorf("Expected README details for %s, got empty string", tc.name)
		}
		if !strings.Contains(obj.Details, "# Test Module") {
			t.Errorf("Expected README content in details for %s", tc.name)
		}
	}

	// Clean up output.json file created by the function
	outputFile := "output.json"
	if _, err := os.Stat(outputFile); err == nil {
		os.Remove(outputFile)
	}
}

func TestProcessBatchOutput_MissingFiles(t *testing.T) {
	// Create a temporary test directory structure with minimal files
	tempDir := t.TempDir()

	// Create subdirectories
	combinedVarDir := filepath.Join(tempDir, "combinedVar")
	examplesDir := filepath.Join(tempDir, "examples")
	readmesDir := filepath.Join(tempDir, "readmes")

	err := os.MkdirAll(combinedVarDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(examplesDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(readmesDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create only a variables file, no example or README
	variableHCL := `variable "test_var" {
  type        = string
  description = "A test variable."
  nullable    = false
}`

	variablesFilePath := filepath.Join(combinedVarDir, "missing-files-test_variables.tf")
	err = os.WriteFile(variablesFilePath, []byte(variableHCL), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test ProcessBatchOutput with missing files
	result, err := ProcessBatchOutput(tempDir)
	if err != nil {
		t.Fatalf("ProcessBatchOutput failed: %v", err)
	}

	// Verify results
	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	if len(result) > 0 {
		obj := result[0]

		// Check that example HCL is empty (file not found)
		if obj.ExampleHCL != "" {
			t.Errorf("Expected empty ExampleHCL when file missing, got: %s", obj.ExampleHCL)
		}

		// Check that details are empty (README not found)
		if obj.Details != "" {
			t.Errorf("Expected empty Details when README missing, got: %s", obj.Details)
		}

		// Check that fields still exist (from variables file)
		if len(obj.Fields) == 0 {
			t.Errorf("Expected fields from variables file, got none")
		}
	}

	// Clean up output.json file created by the function
	outputFile := "output.json"
	if _, err := os.Stat(outputFile); err == nil {
		os.Remove(outputFile)
	}
}

func TestProcessBatchOutput_SourceReplacement(t *testing.T) {
	// Test different source path formats and spacing variations
	tempDir := t.TempDir()

	combinedVarDir := filepath.Join(tempDir, "combinedVar")
	examplesDir := filepath.Join(tempDir, "examples")
	readmesDir := filepath.Join(tempDir, "readmes")

	for _, dir := range []string{combinedVarDir, examplesDir, readmesDir} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test cases with different spacing patterns
	testCases := []struct {
		name       string
		sourceText string
		expected   string
	}{
		{
			name:       "normal-spacing",
			sourceText: `source = "../../"`,
			expected:   `source = "Azure/source-test/azurerm"`,
		},
		{
			name:       "extra-spaces",
			sourceText: `source   =   "../../"`,
			expected:   `source = "Azure/source-test/azurerm"`,
		},
		{
			name:       "tabs-and-spaces",
			sourceText: "source\t=\t\"../../\"",
			expected:   `source = "Azure/source-test/azurerm"`,
		},
		{
			name:       "no-spaces",
			sourceText: `source="../../"`,
			expected:   `source = "Azure/source-test/azurerm"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create variables file
			variableHCL := `variable "test_var" {
  type        = string
  description = "A test variable."
  nullable    = false
}`

			variablesFilePath := filepath.Join(combinedVarDir, "source-test_variables.tf")
			err := os.WriteFile(variablesFilePath, []byte(variableHCL), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Create example file with specific source format
			exampleHCL := fmt.Sprintf(`module "test" {
  %s
  
  test_var = "value"
}`, tc.sourceText)

			exampleFilePath := filepath.Join(examplesDir, "source-test_example.tf")
			err = os.WriteFile(exampleFilePath, []byte(exampleHCL), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Process the files
			result, err := ProcessBatchOutput(tempDir)
			if err != nil {
				t.Fatalf("ProcessBatchOutput failed: %v", err)
			}

			// Verify source replacement
			if len(result) > 0 {
				if !strings.Contains(result[0].ExampleHCL, tc.expected) {
					t.Errorf("Expected '%s' in result, but got: %s", tc.expected, result[0].ExampleHCL)
				}
			}

			// Clean up for next iteration
			os.Remove(variablesFilePath)
			os.Remove(exampleFilePath)
		})
	}

	// Clean up output.json file created by the function
	outputFile := "output.json"
	if _, err := os.Stat(outputFile); err == nil {
		os.Remove(outputFile)
	}
}

func TestProcessBatchOutput_JSONOutput(t *testing.T) {
	// Test that the output JSON is valid and contains expected structure
	tempDir := t.TempDir()

	combinedVarDir := filepath.Join(tempDir, "combinedVar")
	examplesDir := filepath.Join(tempDir, "examples")
	readmesDir := filepath.Join(tempDir, "readmes")

	for _, dir := range []string{combinedVarDir, examplesDir, readmesDir} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create a simple test case
	variableHCL := `variable "test_var" {
  type        = string
  description = "A test variable."
  nullable    = false
}`

	variablesFilePath := filepath.Join(combinedVarDir, "json-test_variables.tf")
	err := os.WriteFile(variablesFilePath, []byte(variableHCL), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Process the files
	result, err := ProcessBatchOutput(tempDir)
	if err != nil {
		t.Fatalf("ProcessBatchOutput failed: %v", err)
	}

	// Check that output.json was created and is valid JSON
	outputFile := "output.json"
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Expected output.json to be created")
	} else {
		// Read and parse the JSON
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output.json: %v", err)
		}

		var jsonResult []schema.TerraformObject
		err = json.Unmarshal(content, &jsonResult)
		if err != nil {
			t.Fatalf("Failed to parse output.json: %v", err)
		}

		// Compare with function result
		if len(jsonResult) != len(result) {
			t.Errorf("JSON output length %d doesn't match function result length %d", len(jsonResult), len(result))
		}

		// Clean up
		os.Remove(outputFile)
	}
}
