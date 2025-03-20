package processors

import (
	"fmt"

	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
	"github.com/Azure/azurerm-lsp/provider-schema/processors/.tools/document-lint/model"
)

// CombineSchemaAndMarkdown merges markdown fields into the schema and returns the combined resources.
func CombineSchemaAndMarkdown(providerSchema *schema.ProviderSchema, markdownDocs map[string]*model.ResourceDoc) (TerraformObjects, error) {
	terraformObjects := make(TerraformObjects)

	// Combine resources and data sources
	for name, resourceSchema := range providerSchema.ResourceSchemas {
		terraformObject := &TerraformObject{
			Name:   name,
			Fields: make(map[string]*schema.SchemaAttribute),
		}
		terraformObjects[name] = terraformObject

		markdownDoc, exists := markdownDocs[name]
		if !exists || markdownDoc == nil {
			//fmt.Println("Resource/DataSource not found in documentation:", name)
			continue
		}

		terraformObject.ExampleHCL = markdownDoc.ExampleHCL
		terraformObject.Timeouts = markdownDoc.Timeouts
		terraformObject.Import = markdownDoc.Import

		// Inject descriptions from markdown into the providerSchema fields
		combineFieldsRecursively(name, resourceSchema.Block, markdownDoc.AllProp(), terraformObject.Fields, "")
	}

	return terraformObjects, nil
}

// combineFieldsRecursively recursively combines fields from the schema and markdown properties.
func combineFieldsRecursively(resourceName string, schemaBlock *schema.SchemaBlock, markdownProps model.Properties, fields map[string]*schema.SchemaAttribute, attributePath string) {
	// Inject descriptions for attributes (fields)
	for fieldName, schemaField := range schemaBlock.Attributes {
		schemaField.Name = fieldName
		schemaField.ResourceOrDataSourceName = resourceName
		schemaField.AttributePath = buildAttributePath(attributePath, fieldName)

		if markdownProps != nil {
			markdownField := markdownProps[fieldName]

			if markdownField != nil {
				schemaField.Description = markdownField.Content
				schemaField.PossibleValues = markdownField.PossibleValues()
			} else {
				fmt.Printf("(TerraformObject %s) Field not found in documentation: %s\n", resourceName, schemaField.AttributePath)
			}
		}

		fields[schemaField.Name] = schemaField
	}

	// Inject descriptions for nested blocks
	for blockName, schemaBlockType := range schemaBlock.NestedBlocks {
		combinedBlockField := &schema.SchemaAttribute{
			Name:                     blockName,
			AttributeType:            schemaBlockType.Block.ImpliedType(),
			Required:                 schemaBlockType.Required,
			Optional:                 schemaBlockType.Optional,
			Computed:                 schemaBlockType.Computed,
			ConflictsWith:            schemaBlockType.ConflictsWith,
			ExactlyOneOf:             schemaBlockType.ExactlyOneOf,
			AtLeastOneOf:             schemaBlockType.AtLeastOneOf,
			RequiredWith:             schemaBlockType.RequiredWith,
			ResourceOrDataSourceName: resourceName,
			AttributePath:            buildAttributePath(attributePath, blockName),
			NestingMode:              schemaBlockType.NestingMode,
			Fields:                   make(map[string]*schema.SchemaAttribute),
		}

		// Inject description from markdown if available
		if markdownProps != nil {
			markdownBlockProps := markdownProps.FindAllSubBlock(blockName)
			if len(markdownBlockProps) > 0 {
				// Use the first field's description as the block description
				combinedBlockField.Description = markdownBlockProps[0].Content
				combinedBlockField.PossibleValues = markdownBlockProps[0].PossibleValues()

				// Inject descriptions for nested attributes
				combineFieldsRecursively(resourceName, schemaBlockType.Block, markdownBlockProps[0].Subs, combinedBlockField.Fields, combinedBlockField.AttributePath)
			} else {
				fmt.Printf("(TerraformObject %s) Block not found in documentation: %s\n", resourceName, combinedBlockField.AttributePath)
			}
		}

		fields[combinedBlockField.Name] = combinedBlockField
	}
}

// buildAttributePath constructs the attribute path for a field in format "parent.field".
func buildAttributePath(parentPath string, fieldName string) string {
	if parentPath == "" {
		return fieldName
	}

	return fmt.Sprintf("%s.%s", parentPath, fieldName)
}
