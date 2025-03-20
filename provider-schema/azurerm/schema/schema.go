package schema

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// This is a simplified and modified version of the hashicorp/terraform-json.
// The motivation for this is to add more information that is lost during the conversion from plugin sdk (v2) to the terraform core schema.
// (github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema/core_schema.go)
// Specifically, we are:
// 1. adding Required, Optional, Computed for the SchemaBlockType
// 2. adding Default for the SchemaAttribute
// 3. adding ExactlyOneOf, AtLeastOneOf, ConflictsWith and RequiredWith for both SchemaBlockType and the SchemaAttribute
// 4. removing any other attributes

type ProviderSchema struct {
	ResourceSchemas map[string]*Schema `json:"resource_schemas,omitempty"`
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
	Computed bool `json:"computed,omitempty"`

	ConflictsWith []string `json:"conflicts_with,omitempty"`
	ExactlyOneOf  []string `json:"exactly_one_of,omitempty"`
	AtLeastOneOf  []string `json:"at_least_one_of,omitempty"`
	RequiredWith  []string `json:"required_with,omitempty"`
}

type SchemaAttribute struct {
	Name          string   `json:"name,omitempty"`
	AttributeType cty.Type `json:"type,omitempty"`

	Required bool        `json:"required,omitempty"`
	Optional bool        `json:"optional,omitempty"`
	Computed bool        `json:"computed,omitempty"`
	Default  interface{} `json:"default,omitempty"`

	ConflictsWith []string `json:"conflicts_with,omitempty"`
	ExactlyOneOf  []string `json:"exactly_one_of,omitempty"`
	AtLeastOneOf  []string `json:"at_least_one_of,omitempty"`
	RequiredWith  []string `json:"required_with,omitempty"`

	// The following fields are not part of the original schema but are added for ease of use
	ResourceOrDataSourceName string `json:"resource_or_data_source_name,omitempty"`
	AttributePath            string `json:"attribute_path,omitempty"`
	Description              string `json:"description,omitempty"`

	PossibleValues []string `json:"possible_values,omitempty"`

	// Block specifics
	NestingMode NestingMode                 `json:"nesting_mode,omitempty"`
	Fields      map[string]*SchemaAttribute `json:"fields,omitempty"`
}

func (b *SchemaAttribute) GetDescription() string {
	if b.Description == "" {
		return "UnDocumented"
	}

	return b.Description
}

func (b *SchemaAttribute) GetAttributeDocLink(parentLink string) string {
	fieldParts := strings.Split(b.AttributePath, ".")
	if len(fieldParts) == 0 {
		return parentLink
	}

	outerField := fieldParts[0]
	return fmt.Sprintf("%s#%s", parentLink, outerField)
}
