package schema

import (
	"fmt"
	"strings"

	"github.com/Azure/ms-terraform-lsp/internal/azure/types"
	"github.com/Azure/ms-terraform-lsp/internal/parser"
)

func GetAzAPIDef(resourceType *types.TypeBase, hclNodes []*parser.HclNode, index int) []*types.TypeBase {
	if resourceType == nil {
		return nil
	}
	if len(hclNodes) == index {
		if t, ok := (*resourceType).(*types.DiscriminatedObjectType); ok {
			if discriminator, ok := hclNodes[index-1].Children[t.Discriminator]; ok && discriminator != nil && discriminator.Value != nil {
				if discriminatorValue := strings.Trim(*discriminator.Value, `"`); len(discriminatorValue) > 0 {
					if t.Elements[discriminatorValue] != nil && t.Elements[discriminatorValue].Type != nil {
						selectedDiscriminatedObjectType := &types.DiscriminatedObjectType{
							Name:           t.Name,
							Discriminator:  t.Discriminator,
							BaseProperties: t.BaseProperties,
							Elements: map[string]*types.TypeReference{
								discriminatorValue: t.Elements[discriminatorValue],
							},
						}
						return []*types.TypeBase{selectedDiscriminatedObjectType.AsTypeBase()}
					}
				}
			}
		}
		return []*types.TypeBase{resourceType}
	}
	key := hclNodes[index].Key
	switch t := (*resourceType).(type) {
	case *types.ArrayType:
		if t.ItemType != nil {
			if strings.Contains(key, ".") {
				return GetAzAPIDef(t.ItemType.Type, hclNodes, index+1)
			}
			return GetAzAPIDef(t.ItemType.Type, hclNodes, index)
		}
		return nil
	case *types.DiscriminatedObjectType:
		if value, ok := t.BaseProperties[key]; ok {
			if value.Type != nil {
				return GetAzAPIDef(value.Type.Type, hclNodes, index+1)
			}
		}
		if index != 0 {
			if discriminator, ok := hclNodes[index-1].Children[t.Discriminator]; ok && discriminator != nil && discriminator.Value != nil {
				if discriminatorValue := strings.Trim(*discriminator.Value, `"`); len(discriminatorValue) > 0 {
					if t.Elements[discriminatorValue] != nil && t.Elements[discriminatorValue].Type != nil {
						return GetAzAPIDef(t.Elements[discriminatorValue].Type, hclNodes, index)
					}
				}
			}
		}
		res := make([]*types.TypeBase, 0)
		for _, discriminator := range t.Elements {
			if resourceType := GetAzAPIDef(discriminator.Type, hclNodes, index); resourceType != nil {
				res = append(res, resourceType...)
			}
		}
		return res
	case *types.ObjectType:
		if value, ok := t.Properties[key]; ok {
			if value.Type != nil {
				return GetAzAPIDef(value.Type.Type, hclNodes, index+1)
			}
		}
		if t.AdditionalProperties != nil {
			return GetAzAPIDef(t.AdditionalProperties.Type, hclNodes, index+1)
		}
	case *types.ResourceType:
		if t.Body != nil {
			return GetAzAPIDef(t.Body.Type, hclNodes, index+1)
		}
	case *types.ResourceFunctionType:
		if t.Input != nil {
			return GetAzAPIDef(t.Input.Type, hclNodes, index+1)
		}
	case *types.AnyType:
		return []*types.TypeBase{resourceType}
	case *types.BooleanType:
		return []*types.TypeBase{resourceType}
	case *types.IntegerType:
		return []*types.TypeBase{resourceType}
	case *types.StringType:
		return []*types.TypeBase{resourceType}
	case *types.StringLiteralType:
		return []*types.TypeBase{resourceType}
	case *types.UnionType:
		res := make([]*types.TypeBase, 0)
		for _, element := range t.Elements {
			res = append(res, GetAzAPIDef(element.Type, hclNodes, index)...)
		}
		return res
	}
	return nil
}

func GetAzAPIAllowedProperties(resourceType *types.TypeBase) []Property {
	if resourceType == nil {
		return []Property{}
	}
	props := make([]Property, 0)
	switch t := (*resourceType).(type) {
	case *types.ArrayType:
		return props
	case *types.DiscriminatedObjectType:
		for key, value := range t.BaseProperties {
			if prop := AzAPIPropertyFromObjectProperty(key, value); prop != nil {
				props = append(props, *prop)
			}
		}
		for _, discriminator := range t.Elements {
			props = append(props, GetAzAPIAllowedProperties(discriminator.Type)...)
		}
	case *types.ObjectType:
		for key, value := range t.Properties {
			if prop := AzAPIPropertyFromObjectProperty(key, value); prop != nil {
				props = append(props, *prop)
			}
		}
		if t.AdditionalProperties != nil {
			props = append(props, GetAzAPIAllowedProperties(t.AdditionalProperties.Type)...)
		}
	case *types.ResourceType:
		if t.Body != nil {
			return GetAzAPIAllowedProperties(t.Body.Type)
		}
	case *types.AnyType:
	case *types.BooleanType:
	case *types.IntegerType:
	case *types.StringType:
	case *types.StringLiteralType:
	case *types.UnionType:
	}
	return props
}

func GetAzAPIAllowedValues(resourceType *types.TypeBase) []string {
	if resourceType == nil {
		return nil
	}
	values := make([]string, 0)
	switch t := (*resourceType).(type) {
	case *types.ResourceType:
		if t.Body != nil {
			return GetAzAPIAllowedValues(t.Body.Type)
		}
	case *types.StringLiteralType:
		return []string{fmt.Sprintf(`"%s"`, t.Value)}
	case *types.UnionType:
		for _, element := range t.Elements {
			values = append(values, GetAzAPIAllowedValues(element.Type)...)
		}
		return values
	case *types.DiscriminatedObjectType:
	case *types.ObjectType:
	case *types.ArrayType:
	case *types.AnyType:
	case *types.BooleanType:
		values = append(values, "true", "false")
	case *types.IntegerType:
	case *types.StringType:
	}
	return values
}

func AzAPIPropertyFromObjectProperty(propertyName string, property types.ObjectProperty) *Property {
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
		propertyType = GetAzAPITypeName(property.Type.Type)
	}
	return &Property{
		Name:        propertyName,
		Description: description,
		Modifier:    modifier,
		Type:        propertyType,
	}
}

func GetAzAPITypeName(typeBase *types.TypeBase) string {
	if typeBase == nil {
		return ""
	}
	switch t := (*typeBase).(type) {
	case *types.ArrayType:
		return "array"
	case *types.DiscriminatedObjectType:
		return "object"
	case *types.ObjectType:
		return "object"
	case *types.ResourceType:
		return "object"
	case *types.AnyType:
		return "any"
	case *types.BooleanType:
		return "boolean"
	case *types.IntegerType:
		return "int"
	case *types.StringType:
		return "string"
	case *types.StringLiteralType:
		return "string"
	case *types.UnionType:
		for _, element := range t.Elements {
			return GetAzAPITypeName(element.Type)
		}
	}
	return ""
}

func GetAzAPIRequiredPropertySet(typeBase *types.TypeBase) []PropertySet {
	if typeBase == nil {
		return nil
	}
	switch t := (*typeBase).(type) {
	case *types.DiscriminatedObjectType:
		res := make([]PropertySet, 0)
		for name, element := range t.Elements {
			if element == nil {
				continue
			}
			propertySet := GetAzAPIRequiredPropertySet(element.Type)
			if len(propertySet) == 1 {
				requiredProps := propertySet[0].Properties
				requiredProps[t.Discriminator] = Property{
					Name:  t.Discriminator,
					Value: name,
				}

				res = append(res, PropertySet{
					Name:       name,
					Properties: requiredProps,
				})
			}
		}
		return res
	case *types.ObjectType:
		requiredProps := make(map[string]Property)
		for propName, prop := range t.Properties {
			if prop.IsRequired() {
				if value := AzAPIPropertyFromObjectProperty(propName, prop); value != nil {
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
			return GetAzAPIRequiredPropertySet(t.Body.Type)
		}
	case *types.UnionType:
		res := make([]PropertySet, 0)
		for _, element := range t.Elements {
			if element.Type == nil {
				continue
			}
			res = append(res, GetAzAPIRequiredPropertySet(element.Type)...)
		}
		return res
	}
	return nil
}
