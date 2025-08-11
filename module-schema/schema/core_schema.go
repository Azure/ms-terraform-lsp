package schema

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

const (
	variable    = "variable"
	hclType     = "type"
	description = "description"
	nullable    = "nullable"
	hclDefault  = "default"
	sensitive   = "sensitive"
	validation  = "validation"
)

func FromHCLFile(fileName string) (*SchemaBlock, error) {
	var block SchemaBlock
	parser := hclparse.NewParser()
	hclFile, diag := parser.ParseHCLFile(fileName)
	if diag.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %v", diag)
	}

	schemaAttr, error := FromHCLBody(hclFile.Body)
	if error != nil {
		return nil, fmt.Errorf("failed to convert HCL body to schema: %v", error)
	}
	block.Attributes = schemaAttr
	return &block, nil
}

func FromHCLBody(body hcl.Body) (map[string]*SchemaAttribute, error) {
	res := make(map[string]*SchemaAttribute)
	content, diag := body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"variable"},
			},
		},
	})
	if diag.HasErrors() {
		return nil, fmt.Errorf("failed to get content from HCL body: %v\n", diag)
	}

	for _, block := range content.Blocks {
		var varName string
		if block.Type == variable {
			varName = block.Labels[0]
		}

		blockBody := block.Body
		blockContent, diag := blockBody.Content(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: hclType},
				{Name: description},
				{Name: nullable},
				{Name: hclDefault},
				{Name: sensitive},
			},
			Blocks: []hcl.BlockHeaderSchema{
				{Type: validation},
			},
		})
		if diag.HasErrors() {
			return nil, fmt.Errorf("failed to get block content: %v\n", diag)
		}

		attribute, err := fromHCLAttributes(blockContent.Attributes, varName)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attributes: %v\n", err)
		}

		res[varName] = attribute
	}

	return res, nil
}

func fromHCLAttributes(attrs hcl.Attributes, varName string) (*SchemaAttribute, error) {
	result := &SchemaAttribute{
		Name:          varName,
		AttributePath: varName,
		Optional:      true,
	}
	var defaultValueFromType cty.Value
	for name, attr := range attrs {
		switch attr.Name {
		case hclType:
			attrType, v, err := fromHCLExpression(attr.Expr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse type for attribute %s: %v\n", name, err)
			}
			result.AttributeType = *attrType
			if v != nil {
				defaultValueFromType = *v
			}
		case description:
			val, diags := attr.Expr.Value(&hcl.EvalContext{})
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to parse description for attribute %s: %v\n", name, diags)
			}
			result.Content = val.AsString()
		case nullable:
			val, diags := attr.Expr.Value(&hcl.EvalContext{})
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to parse nullable for attribute %s: %v\n", name, diags)
			}
			result.Optional = val.True()
		case hclDefault:
			val, diags := attr.Expr.Value(&hcl.EvalContext{})
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to parse default for attribute %s: %v\n", name, diags)
			}

			if val.IsNull() {
				result.Default = nil
				continue
			}

			// For primitive types, unwrap the cty.Value to a standard Go type.
			// For complex types (object, map, list, etc.), keep it as a cty.Value.
			switch val.Type() {
			case cty.String:
				result.Default = val.AsString()
			case cty.Number:
				result.Default = val.AsBigFloat()
			case cty.Bool:
				result.Default = val.True()
			default:
				result.Default = val
			}
		default:
			// unknown attribute
		}
	}
	mergedDefaultValue := mergeDefaultValue(defaultValueFromType, result.Default)
	if mergedDefaultValue != nil {
		// If the merged value is a cty.Value, check if it's null.
		if ctyVal, ok := mergedDefaultValue.(cty.Value); ok && ctyVal.IsNull() {
			result.Default = nil
		} else {
			result.Default = mergedDefaultValue
		}
	} else if result.AttributeType != cty.NilType {
		result.Default = nil
	}

	if result.Default != nil && result.AttributeType.Equals(cty.String) {
		var defaultString string
		isString := false
		if str, ok := result.Default.(string); ok {
			defaultString = str
			isString = true
		} else if ctyVal, ok := result.Default.(cty.Value); ok && ctyVal.Type() == cty.String {
			defaultString = ctyVal.AsString()
			isString = true
			result.Default = defaultString
		}

		if isString {
			result.PossibleValues = []string{defaultString}
		}
	}

	return result, nil
}

func fromHCLExpression(expr hcl.Expression) (*cty.Type, *cty.Value, error) {
	typeStruc, defaultValue, diag := typeexpr.TypeConstraintWithDefaults(expr)
	if diag.HasErrors() {
		va, _ := expr.Value(&hcl.EvalContext{})
		if va.IsNull() {
			return nil, nil, nil
		} else if va.Type() == cty.Bool {
			return &cty.Bool, &va, nil
		} else if va.Type() == cty.String {
			return &cty.String, &va, nil
		} else if va.Type() == cty.Number {
			return &cty.Number, &va, nil
		} else if va.Type().IsObjectType() {
			obj := cty.Object(map[string]cty.Type{})
			return &obj, &va, nil
		} else if va.Type().IsTupleType() {
			tuple := cty.Tuple([]cty.Type{})
			return &tuple, &va, nil
		}
		return nil, nil, fmt.Errorf("failed to parse type expression: %v\n", diag)
	}
	va, _ := fromHCLDefaults(defaultValue)
	return &typeStruc, va, nil
}

func fromHCLDefaults(defaults *typeexpr.Defaults) (*cty.Value, error) {
	if defaults == nil {
		return nil, fmt.Errorf("defaults is nil")
	}
	var res cty.Value
	defaultType := defaults.Type
	if defaultType.IsObjectType() {
		attrs := make(map[string]cty.Value)
		for name, val := range defaults.DefaultValues {
			attrs[name] = val
		}
		res = cty.ObjectVal(attrs)
	} else if defaultType.IsMapType() && defaultType.ElementType().IsObjectType() {
		children := defaults.Children
		childMap := make(map[string]cty.Value)
		for name, child := range children {
			childValue, err := fromHCLDefaults(child)
			if err != nil {
				return nil, fmt.Errorf("failed to parse child defaults for %s: %v\n", name, err)
			}
			childMap[name] = *childValue
		}
		res = cty.MapVal(childMap)
	} else {
		if len(defaults.DefaultValues) > 0 {
			for _, val := range defaults.DefaultValues {
				res = val
				break
			}
		} else {
			res = cty.NullVal(defaultType)
		}
	}
	return &res, nil
}

func mergeDefaultValue(defaultFromType cty.Value, currentDefault any) any {
	// If the main 'default' attribute is set (and not null), it always takes precedence.
	if currentDefault != nil {
		if val, ok := currentDefault.(cty.Value); ok && !val.IsNull() {
			return currentDefault
		}
		// Handle primitives that might have been unwrapped
		if _, ok := currentDefault.(cty.Value); !ok {
			return currentDefault
		}
	}

	// Otherwise, use the default value derived from the 'type' expression, if it's valid.
	if !defaultFromType.IsNull() {
		return defaultFromType
	}

	return nil // No valid default found
}

func mergeObjects(defaultMap, currentMap map[string]cty.Value) cty.Value {
	merged := buildValueRecursively(defaultMap)
	modifyValueRecursively(merged, currentMap)
	return cty.ObjectVal(merged)
}

func buildValueRecursively(input map[string]cty.Value) map[string]cty.Value {
	res := make(map[string]cty.Value)
	for key, value := range input {
		if value.Type().IsObjectType() {
			if value.IsNull() {
				continue
			}
			vMap := buildValueRecursively(value.AsValueMap())
			res[key] = cty.ObjectVal(vMap)
		} else {
			res[key] = value
		}
	}
	return res
}

func modifyValueRecursively(res, input map[string]cty.Value) {
	for key, value := range input {
		if existingValue, ok := res[key]; ok {
			if existingValue.Type().IsObjectType() && value.Type().IsObjectType() {
				existingMap := existingValue.AsValueMap()
				valueMap := value.AsValueMap()
				modifyValueRecursively(existingMap, valueMap)
				res[key] = cty.ObjectVal(existingMap)
			} else if existingValue.Type().IsMapType() && value.Type().IsMapType() {
				existingMap := existingValue.AsValueMap()
				valueMap := value.AsValueMap()
				modifyValueRecursively(existingMap, valueMap)
				res[key] = cty.MapVal(existingMap)
			} else {
				res[key] = value
			}
		} else {
			res[key] = value
		}
	}
}
