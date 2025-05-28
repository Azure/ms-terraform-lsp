package schema

import (
	"fmt"
	"strings"

	"github.com/microsoft/msgraph-lsp/internal/parser"
	"github.com/ms-henglu/go-msgraph-types/types"
)

func GetDef(resourceType *types.TypeBase, hclNodes []*parser.HclNode, index int) []*types.TypeBase {
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
				return GetDef(&t.ItemType.Type, hclNodes, index+1)
			}
			return GetDef(&t.ItemType.Type, hclNodes, index)
		}
		return nil
	case *types.ObjectType:
		if value, ok := t.Properties[key]; ok {
			if value.Type != nil {
				return GetDef(&value.Type.Type, hclNodes, index+1)
			}
		}
		if t.AdditionalProperties != nil {
			return GetDef(&t.AdditionalProperties.Type, hclNodes, index+1)
		}
	case *types.ResourceType:
		if t.Body != nil {
			return GetDef(&t.Body.Type, hclNodes, index+1)
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
			res = append(res, GetDef(&element.Type, hclNodes, index)...)
		}
		return res
	}
	return nil
}

func GetAllowedProperties(resourceType *types.TypeBase) []Property {
	if resourceType == nil {
		return []Property{}
	}
	props := make([]Property, 0)
	switch t := (*resourceType).(type) {
	case *types.ArrayType:
		return props
	case *types.ObjectType:
		for key, value := range t.Properties {
			if prop := PropertyFromObjectProperty(key, value); prop != nil {
				props = append(props, *prop)
			}
		}
		if t.AdditionalProperties != nil {
			props = append(props, GetAllowedProperties(&t.AdditionalProperties.Type)...)
		}
	case *types.ResourceType:
		if t.Body != nil {
			return GetAllowedProperties(&t.Body.Type)
		}
	case *types.AnyType:
	case *types.BooleanType:
	case *types.NumberType:
	case *types.StringType:
	case *types.UnionType:
	}
	return props
}

func GetAllowedValues(resourceType *types.TypeBase) []string {
	if resourceType == nil {
		return nil
	}
	values := make([]string, 0)
	switch t := (*resourceType).(type) {
	case *types.ResourceType:
		if t.Body != nil {
			return GetAllowedValues(&t.Body.Type)
		}
	case *types.UnionType:
		for _, element := range t.Elements {
			values = append(values, GetAllowedValues(&element.Type)...)
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

func PropertyFromObjectProperty(propertyName string, property types.ObjectProperty) *Property {
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
		propertyType = GetTypeName(&property.Type.Type)
	}
	return &Property{
		Name:        propertyName,
		Description: description,
		Modifier:    modifier,
		Type:        propertyType,
	}
}

func GetTypeName(typeBase *types.TypeBase) string {
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
			return GetTypeName(&element.Type)
		}
	}
	return ""
}
