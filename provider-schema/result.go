package provider_schema

import (
	"fmt"
	"strings"

	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
	"github.com/Azure/azurerm-lsp/provider-schema/processors"
)

func Run() (processors.TerraformObjects, error) {
	providerPath := "/Users/harryqu/Projects-m/terraform-m"
	gitBranch := "main"
	outputFile := "combined_output.json"

	return processors.ProcessOutput(providerPath, gitBranch, outputFile)
}

var finalTerraformObject processors.TerraformObjects

func GetFinalTerraformObject() processors.TerraformObjects {
	if finalTerraformObject == nil {
		terraformObject, err := Run()
		if err != nil {
			panic(err)
		}

		finalTerraformObject = terraformObject
	}

	return finalTerraformObject
}

func ListAllResources() []string {
	var resources []string

	for name, terraformObject := range GetFinalTerraformObject() {
		if terraformObject.IsDataSource() {
			continue
		}

		resources = append(resources, name)
	}

	return resources
}

func ListAllDataSources() []string {
	var dataSources []string

	for _, terraformObject := range GetFinalTerraformObject() {
		if !terraformObject.IsDataSource() {
			continue
		}

		dataSources = append(dataSources, terraformObject.GetName())
	}

	return dataSources
}

func ListAllResourcesAndDataSources() []string {
	var resourcesAndDataSources []string

	for name := range GetFinalTerraformObject() {
		resourcesAndDataSources = append(resourcesAndDataSources, name)
	}

	return resourcesAndDataSources
}

// NavigateToNestedBlock navigates down a nested block path and returns the final block or property
func NavigateToNestedBlock(objectName, path string) (*schema.SchemaAttribute, error) {
	resource, exists := GetFinalTerraformObject()[objectName]
	if !exists {
		return nil, fmt.Errorf("resource/data source '%s' not found", objectName)
	}

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path '%s'", path)
	}

	curFields := resource.Fields
	var result *schema.SchemaAttribute
	for _, part := range parts {
		result, exists = curFields[part]
		if !exists {
			return nil, fmt.Errorf("path '%s' not found in resource/data source '%s'", path, objectName)
		}
		curFields = result.Fields
	}

	if result == nil {
		return nil, fmt.Errorf("path '%s' is nil in resource/data source '%s'", path, objectName)
	}

	return result, nil
}

func GetPropertyInfo(objectName, propertyPath string) (*schema.SchemaAttribute, error) {
	return NavigateToNestedBlock(objectName, propertyPath)
}

func GetPossibleValuesForProperty(objectName, propertyName string) ([]string, error) {
	block, err := NavigateToNestedBlock(objectName, propertyName)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, fmt.Errorf("block '%s' not found in resource/data source '%s'", propertyName, objectName)
	}

	if block.PossibleValues == nil {
		return nil, fmt.Errorf("no possible values found for block '%s' in resource/data source '%s'", propertyName, objectName)
	}

	return block.PossibleValues, nil
}

func ListDirectProperties(objectName string) ([]*schema.SchemaAttribute, error) {
	resource, exists := GetFinalTerraformObject()[objectName]
	if !exists {
		return nil, fmt.Errorf("resource/data source '%s' not found", objectName)
	}

	var properties []*schema.SchemaAttribute
	for _, property := range resource.Fields {
		properties = append(properties, property)
	}

	return properties, nil
}

func ListDirectPropertiesForBlockPath(objectName, blockPath string) ([]*schema.SchemaAttribute, error) {
	block, err := NavigateToNestedBlock(objectName, blockPath)
	if err != nil {
		return nil, err
	}

	var properties []*schema.SchemaAttribute
	for _, property := range block.Fields {
		properties = append(properties, property)
	}

	return properties, nil
}

func GetSnippet(objectName string) (string, error) {
	resource, exists := GetFinalTerraformObject()[objectName]
	if !exists {
		return "", fmt.Errorf("resource/data source '%s' not found", objectName)
	}

	return resource.GetSnippet(), nil
}

func GetResourceOrDataSourceDocLink(objectName string) string {
	resource, exists := GetFinalTerraformObject()[objectName]
	if !exists {
		return ""
	}

	return resource.GetResourceOrDataSourceDocLink()
}
