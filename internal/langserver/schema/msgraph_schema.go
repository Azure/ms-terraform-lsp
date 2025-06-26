package schema

import (
	"fmt"
	"strings"

	"github.com/Azure/ms-terraform-lsp/internal/parser"
	"github.com/ms-henglu/go-msgraph-types/types"
)

func GetMSGraphDef(resourceType *types.TypeBase, hclNodes []*parser.HclNode, index int) []*types.TypeBase {
	if resourceType == nil {
		return nil
	}
	if len(hclNodes) == index {
		return []*types.TypeBase{resourceType}
	}
	key := hclNodes[index].Key
	switch t := (*resourceType).(type) {
	case *types.ArrayType:
		if t.ItemType != nil {
			if strings.Contains(key, ".") {
				return GetMSGraphDef(&t.ItemType.Type, hclNodes, index+1)
			}
			return GetMSGraphDef(&t.ItemType.Type, hclNodes, index)
		}
		return nil
	case *types.ObjectType:
		if value, ok := t.Properties[key]; ok {
			if value.Type != nil {
				return GetMSGraphDef(&value.Type.Type, hclNodes, index+1)
			}
		}
		if t.AdditionalProperties != nil {
			return GetMSGraphDef(&t.AdditionalProperties.Type, hclNodes, index+1)
		}
	case *types.ResourceType:
		if t.Body != nil {
			return GetMSGraphDef(&t.Body.Type, hclNodes, index+1)
		}
	case *types.AnyType:
		return []*types.TypeBase{resourceType}
	case *types.BooleanType:
		return []*types.TypeBase{resourceType}
	case *types.NumberType:
		return []*types.TypeBase{resourceType}
	case *types.StringType:
		return []*types.TypeBase{resourceType}
	case *types.UnionType:
		res := make([]*types.TypeBase, 0)
		for _, element := range t.Elements {
			res = append(res, GetMSGraphDef(&element.Type, hclNodes, index)...)
		}
		return res
	}
	return nil
}

func GetMSGraphAllowedProperties(resourceType *types.TypeBase) []Property {
	if resourceType == nil {
		return []Property{}
	}
	props := make([]Property, 0)
	switch t := (*resourceType).(type) {
	case *types.ArrayType:
		return props
	case *types.ObjectType:
		for key, value := range t.Properties {
			if prop := MSGraphPropertyFromObjectProperty(key, value); prop != nil {
				props = append(props, *prop)
			}
		}
		if t.AdditionalProperties != nil {
			props = append(props, GetMSGraphAllowedProperties(&t.AdditionalProperties.Type)...)
		}
	case *types.ResourceType:
		if t.Body != nil {
			return GetMSGraphAllowedProperties(&t.Body.Type)
		}
	case *types.AnyType:
	case *types.BooleanType:
	case *types.NumberType:
	case *types.StringType:
	case *types.UnionType:
	}
	return props
}

func GetMSGraphAllowedValues(resourceType *types.TypeBase) []string {
	if resourceType == nil {
		return nil
	}
	values := make([]string, 0)
	switch t := (*resourceType).(type) {
	case *types.ResourceType:
		if t.Body != nil {
			return GetMSGraphAllowedValues(&t.Body.Type)
		}
	case *types.UnionType:
		for _, element := range t.Elements {
			values = append(values, GetMSGraphAllowedValues(&element.Type)...)
		}
		return values
	case *types.ObjectType:
	case *types.ArrayType:
	case *types.AnyType:
	case *types.BooleanType:
		values = append(values, "true", "false")
	case *types.StringType:
		for _, v := range t.Enum {
			values = append(values, fmt.Sprintf(`"%s"`, v))
		}
	}
	return values
}

func MSGraphPropertyFromObjectProperty(propertyName string, property types.ObjectProperty) *Property {
	if property.IsReadOnly() {
		return nil
	}
	description := ""
	if property.Description != nil {
		description = *property.Description
	}
	modifier := Optional
	if property.IsRequired() {
		modifier = Required
	}
	propertyType := ""
	if property.Type != nil {
		propertyType = GetMSGraphTypeName(&property.Type.Type)
	}
	return &Property{
		Name:        propertyName,
		Description: description,
		Modifier:    modifier,
		Type:        propertyType,
	}
}

func GetMSGraphTypeName(typeBase *types.TypeBase) string {
	if typeBase == nil {
		return ""
	}
	switch t := (*typeBase).(type) {
	case *types.ArrayType:
		return "array"
	case *types.ObjectType:
		return "object"
	case *types.ResourceType:
		return "object"
	case *types.AnyType:
		return "any"
	case *types.BooleanType:
		return "boolean"
	case *types.NumberType:
		return "number"
	case *types.StringType:
		return "string"
	case *types.UnionType:
		for _, element := range t.Elements {
			return GetMSGraphTypeName(&element.Type)
		}
	}
	return ""
}

func GetMSGraphRequiredPropertySet(typeBase *types.TypeBase) []PropertySet {
	if typeBase == nil {
		return nil
	}
	switch t := (*typeBase).(type) {
	case *types.ObjectType:
		requiredProps := make(map[string]Property)
		for propName, prop := range t.Properties {
			if propName == "@odata.type" {
				continue
			}
			if prop.IsRequired() {
				if value := MSGraphPropertyFromObjectProperty(propName, prop); value != nil {
					requiredProps[value.Name] = *value
				}
			}
		}
		return []PropertySet{{
			Name:       t.Name,
			Properties: requiredProps,
		}}
	case *types.ResourceType:
		if t.Body != nil {
			return GetMSGraphRequiredPropertySet(&t.Body.Type)
		}
	case *types.UnionType:
		res := make([]PropertySet, 0)
		for _, element := range t.Elements {
			if element.Type == nil {
				continue
			}
			res = append(res, GetMSGraphRequiredPropertySet(&element.Type)...)
		}
		return res
	}
	return nil
}
