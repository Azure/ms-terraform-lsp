package processor

import (
	"fmt"
	"github.com/Azure/ms-terraform-lsp/module-schema/schema"
	"github.com/zclconf/go-cty/cty"
)

// FindFieldsFromBlock The input block does not contain any fields, the information is inside `type`, so we need to build the fields(as well as the default values) recursively
func FindFieldsFromBlock(block *schema.SchemaBlock, resourceOrDataSourceName string) (*schema.SchemaBlock, error) {
	res := make(map[string]*schema.SchemaAttribute)
	attrs := block.Attributes
	if attrs == nil {
		return nil, fmt.Errorf("block has no attributes")
	}

	for varName, attr := range attrs {
		currentSchema := schema.SchemaAttribute{
			Name:                     varName,
			Content:                  attr.Content,
			AttributePath:            varName,
			ResourceOrDataSourceName: resourceOrDataSourceName,
		}
		if attr.Optional {
			currentSchema.Optional = true
		}
		if attr.Required || !attr.Optional {
			currentSchema.Required = true
		}

		if attr.AttributeType.IsObjectType() {
			currentSchema.AttributeType = attr.AttributeType.WithoutOptionalAttributesDeep()
			fields, err := buildFieldsRecursively(attr.AttributeType, ConvertAttrDefaultToCtyValue(attr.Default), varName, resourceOrDataSourceName)
			if err != nil {
				return nil, fmt.Errorf("building fields recursively for %s: %w", varName, err)
			}
			currentSchema.Fields = fields
		} else if attr.AttributeType.IsCollectionType() {
			currentSchema.AttributeType = attr.AttributeType.WithoutOptionalAttributesDeep()
			fields, err := buildFieldsRecursively(attr.AttributeType.ElementType(), ConvertAttrDefaultToCtyValue(attr.Default), varName, resourceOrDataSourceName)
			if err != nil {
				return nil, fmt.Errorf("building fields recursively for %s: %w", varName, err)
			}
			currentSchema.Fields = fields
		} else {
			currentSchema.AttributeType = attr.AttributeType
			currentSchema.Default = attr.Default
			if attr.AttributeType.Equals(cty.String) && attr.Default != nil {
				ctyVal := ConvertAttrDefaultToCtyValue(attr.Default)
				if ctyVal != nil {
					defaultValue := extractDefaultValue(*ctyVal)
					if defaultStr, ok := defaultValue.(string); ok {
						currentSchema.PossibleValues = append(currentSchema.PossibleValues, defaultStr)
					}
				}
			}
		}
		res[varName] = &currentSchema
	}
	return &schema.SchemaBlock{
		Attributes: res,
	}, nil
}

func buildFieldsRecursively(attrType cty.Type, dfv *cty.Value, parentPath string, resourceOrDataSourceName string) (map[string]*schema.SchemaAttribute, error) {
	if attrType == cty.NilType {
		return nil, fmt.Errorf("attribute type is nil")
	}

	fields := make(map[string]*schema.SchemaAttribute)

	if attrType.IsObjectType() {
		optionalAttributes := attrType.OptionalAttributes()
		for name, subAttrType := range attrType.AttributeTypes() {
			currentPath := fmt.Sprintf("%s.%s", parentPath, name)
			schemaAttr := &schema.SchemaAttribute{
				Name:                     name,
				AttributeType:            subAttrType,
				AttributePath:            currentPath,
				ResourceOrDataSourceName: resourceOrDataSourceName,
			}

			if _, ok := optionalAttributes[name]; ok {
				schemaAttr.Optional = true
			} else {
				schemaAttr.Required = true
			}

			if subAttrType.IsObjectType() || subAttrType.IsMapType() {
				subAttrDfv := GetAttributeWrapper(name, dfv)
				elementType := subAttrType
				if subAttrType.IsMapType() {
					elementType = subAttrType.ElementType()
				}
				subFields, err := buildFieldsRecursively(elementType, subAttrDfv, currentPath, resourceOrDataSourceName)
				if err != nil {
					return nil, err
				}
				schemaAttr.Fields = subFields
			}

			attrDefault := GetAttributeWrapper(name, dfv)
			if attrDefault != nil {
				schemaAttr.Default = extractDefaultValue(*attrDefault)
				if subAttrType.Equals(cty.String) {
					defaultValue := extractDefaultValue(*attrDefault)
					if defaultStr, ok := defaultValue.(string); ok {
						schemaAttr.PossibleValues = append(schemaAttr.PossibleValues, defaultStr)
					}
				}
			}

			fields[name] = schemaAttr
		}
	}

	return fields, nil
}

// extractDefaultValue unwraps a cty.Value to its corresponding Go primitive type if possible.
// For complex types (collections, objects), it returns the original cty.Value.
func extractDefaultValue(val cty.Value) any {
	if !val.IsKnown() || val.IsNull() {
		return nil
	}

	valType := val.Type()
	if valType.IsPrimitiveType() {
		switch valType {
		case cty.String:
			return val.AsString()
		case cty.Number:
			return val.AsBigFloat()
		case cty.Bool:
			return val.True()
		}
	}

	if valType.IsTupleType() || valType.IsListType() || valType.IsSetType() {
		valueSlice := val.AsValueSlice()
		unwrappedSlice := make([]any, len(valueSlice))
		for i, v := range valueSlice {
			unwrappedSlice[i] = extractDefaultValue(v) // Recurse for nested values
		}
		return unwrappedSlice
	}

	// For objects, maps, or other complex types, return the value as is.
	return val
}

func GetAttributeWrapper(name string, dfv *cty.Value) *cty.Value {
	if dfv == nil || dfv.IsNull() {
		return nil
	}
	// Add this check to ensure the value is an object before checking for attributes.
	if !dfv.Type().IsObjectType() {
		return nil
	}
	if !dfv.Type().HasAttribute(name) {
		return nil
	}
	attrDefault := dfv.GetAttr(name)
	if attrDefault.IsNull() {
		return nil
	}
	return &attrDefault
}

func ConvertAttrDefaultToCtyValue(value any) *cty.Value {
	if value == nil {
		return nil
	}

	if ctyVal, ok := value.(cty.Value); ok {
		return &ctyVal
	}

	if str, ok := value.(string); ok {
		res := cty.StringVal(str)
		return &res
	}

	// Fallback for other types if needed, though the panic was specifically about string.
	// This part might need adjustment if other primitive types are stored in `attr.Default`.
	// For now, we'll assume the panic was the only issue.
	if res, ok := value.(cty.Value); ok {
		return &res
	}
	return nil
}
