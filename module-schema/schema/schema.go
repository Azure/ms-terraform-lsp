package schema

import (
	"encoding/json"
	"github.com/zclconf/go-cty/cty"
)

type NestingMode int

const (
	nestingModeInvalid NestingMode = iota
	NestingSingle
	NestingGroup
	NestingList
	NestingSet
	NestingMap
)

type ProviderSchema struct {
	ResourceSchema map[string]*Schema `json:"resource_schemas,omitempty"`
}

type Schema struct {
	Block *SchemaBlock `json:"block,omitempty"`
}

type SchemaBlock struct {
	Attributes   map[string]*SchemaAttribute `json:"attributes,omitempty"`
	NestedBlocks map[string]*SchemaBlockType `json:"block_types,omitempty"`
}

type SchemaBlockType struct {
	NestingMode NestingMode  `json:"nesting_mode,omitempty"`
	Block       *SchemaBlock `json:"block,omitempty"`

	Required bool `json:"required,omitempty"`
	Optional bool `json:"optional,omitempty"`
}

type SchemaAttribute struct {
	Name          string   `json:"name,omitempty"`
	AttributeType cty.Type `json:"type,omitempty"`

	Required      bool        `json:"required,omitempty"`
	Optional      bool        `json:"optional,omitempty"`
	Default       interface{} `json:"default,omitempty"`
	Content       string      `json:"content,omitempty"`
	AttributePath string      `json:"attribute_path,omitempty"`

	PossibleValues           []string                    `json:"possible_values,omitempty"`
	NestingMode              NestingMode                 `json:"nesting_mode,omitempty"`
	Fields                   map[string]*SchemaAttribute `json:"fields,omitempty"`
	ResourceOrDataSourceName string                      `json:"resource_or_data_source_name,omitempty"`
}

func (attr SchemaAttribute) MarshalJSON() ([]byte, error) {
	type Alias SchemaAttribute
	return json.Marshal(&struct {
		Alias
		AttributeType *string `json:"type,omitempty"`
	}{
		Alias:         Alias(attr),
		AttributeType: translateCtyTypeToString(attr.AttributeType),
	})
}

func translateCtyTypeToString(attrType cty.Type) *string {
	if attrType == cty.NilType {
		return nil
	}
	name := attrType.FriendlyName()
	return &name
}
