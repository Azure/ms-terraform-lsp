package types

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"log"
)

type TypeBase interface {
	AsTypeBase() *TypeBase

	FilterConfigurableFields(interface{}) interface{}

	FilterReadOnlyFields(interface{}) interface{}

	Validate(interface{}, string) []error
}

func NewTypeBaseFromOpenAPISchema(input *openapi3.Schema, cache map[*openapi3.Schema]*TypeBase) *TypeBase {
	if input == nil {
		return nil
	}
	if cache[input] != nil {
		return cache[input]
	}

	if input.Discriminator != nil {
		if input.Discriminator.PropertyName != "@odata.type" {
			fmt.Println("unsupported discriminator")
		}
	}

	if len(input.AllOf) != 0 {
		objectType := &ObjectType{
			Type:                 "object",
			Properties:           map[string]ObjectProperty{},
			AdditionalProperties: nil,
			Sensitive:            false,
		}
		cache[input] = objectType.AsTypeBase()

		objectTypeList := make([]*ObjectType, 0)
		for _, schema := range input.AllOf {
			if schema.Value == nil {
				log.Printf("[WARN] schema.Value is nil")
				continue
			}
			childObjectType := NewTypeBaseFromOpenAPISchema(schema.Value, cache)
			if childObjectType == nil {
				log.Printf("[WARN] objectType is nil")
				continue
			}

			objectTypeList = append(objectTypeList, (*childObjectType).(*ObjectType))
		}

		// combine all object types into one
		for _, objType := range objectTypeList {
			for key, value := range objType.Properties {
				objectType.Properties[key] = value
			}
		}
		return objectType.AsTypeBase()
	}

	if len(input.AnyOf) != 0 && input.Discriminator == nil {
		unionType := &UnionType{
			Type:     "union",
			Elements: make([]*TypeReference, 0),
		}
		cache[input] = unionType.AsTypeBase()

		for _, schema := range input.AnyOf {
			if schema.Value == nil {
				log.Printf("[WARN] schema.Value is nil")
				continue
			}
			element := NewTypeBaseFromOpenAPISchema(schema.Value, cache)
			if element == nil {
				log.Printf("[WARN] element is nil")
				continue
			}

			unionType.Elements = append(unionType.Elements, &TypeReference{
				Type: *element,
			})
		}

		return unionType.AsTypeBase()
	}

	if len(input.OneOf) != 0 && input.Discriminator == nil {
		unionType := &UnionType{
			Type:     "union",
			Elements: make([]*TypeReference, 0),
		}
		cache[input] = unionType.AsTypeBase()

		for _, schema := range input.OneOf {
			if schema.Value == nil {
				log.Printf("[WARN] schema.Value is nil")
				continue
			}
			element := NewTypeBaseFromOpenAPISchema(schema.Value, cache)
			if element == nil {
				log.Printf("[WARN] element is nil")
				continue
			}

			unionType.Elements = append(unionType.Elements, &TypeReference{
				Type: *element,
			})
		}

		return unionType.AsTypeBase()
	}

	switch {
	case input.Type.Is("object"):
		t := ObjectType{
			Type:                 "object",
			Name:                 input.Title,
			AdditionalProperties: nil, //TODO
			Sensitive:            false,
		}
		cache[input] = t.AsTypeBase()

		properties := make(map[string]ObjectProperty)

		requiredSet := make(map[string]bool)
		for _, required := range input.Required {
			requiredSet[required] = true
		}

		for key, value := range input.Properties {
			if value == nil || value.Value == nil {
				log.Printf("[WARN] object property value is nil")
				continue
			}

			valueType := NewTypeBaseFromOpenAPISchema(value.Value, cache)
			if valueType == nil {
				log.Printf("[WARN] valueType is nil")
				continue
			}

			flags := make([]ObjectPropertyFlag, 0)
			if requiredSet[key] {
				flags = append(flags, Required)
			}
			if value.Value.ReadOnly {
				flags = append(flags, ReadOnly)
			}
			if value.Value.WriteOnly {
				flags = append(flags, WriteOnly)
			}

			objectProperty := ObjectProperty{
				Type: &TypeReference{
					Type: *valueType,
				},
				Flags:       flags,
				Description: &value.Value.Description,
			}
			properties[key] = objectProperty
		}

		t.Properties = properties
		return t.AsTypeBase()
	case input.Type.Is("string"):
		t := StringType{
			Type:      "string",
			MinLength: &input.MinLength,
			MaxLength: input.MaxLength,
			Sensitive: false,
			Pattern:   input.Pattern,
		}
		if input.Enum != nil {
			t.Enum = make([]string, 0)
			for _, value := range input.Enum {
				t.Enum = append(t.Enum, value.(string))
			}
		}
		cache[input] = t.AsTypeBase()
		return t.AsTypeBase()
	case input.Type.Is("boolean"):
		t := BooleanType{
			Type: "boolean",
		}
		cache[input] = t.AsTypeBase()
		return t.AsTypeBase()
	case input.Type.Is("array"):
		t := ArrayType{
			Type:      "array",
			MinLength: &input.MinItems,
			MaxLength: input.MaxItems,
		}
		cache[input] = t.AsTypeBase()

		var itemType *TypeBase
		if input.Items != nil {
			itemType = NewTypeBaseFromOpenAPISchema(input.Items.Value, cache)
		} else {
			log.Printf("[WARN] array item is nil")
		}

		t.ItemType = &TypeReference{
			Type: *itemType,
		}
		return t.AsTypeBase()
	case input.Type.Is("number"):
		t := NumberType{
			Type:     "number",
			Format:   input.Format,
			MinValue: input.Min,
			MaxValue: input.Max,
		}
		cache[input] = t.AsTypeBase()
		return t.AsTypeBase()
	case input.Type == nil:
		t := AnyType{
			Type:        "any",
			Description: input.Description,
		}
		cache[input] = t.AsTypeBase()
		return t.AsTypeBase()
	default:
		fmt.Println("unsupported type")
	}

	t := AnyType{
		Type:        "any",
		Description: input.Description,
	}
	cache[input] = t.AsTypeBase()
	return t.AsTypeBase()
}
