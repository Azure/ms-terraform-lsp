package types

import (
	"fmt"
)

var _ TypeBase = &NumberType{}

type NumberType struct {
	Type     string   `json:"$type"`
	Format   string   `json:"format"`
	MinValue *float64 `json:"minValue"`
	MaxValue *float64 `json:"maxValue"`
}

func (t *NumberType) Validate(body interface{}, path string) []error {
	if body == nil {
		return nil
	}
	var v int
	switch input := body.(type) {
	case float64, float32:
		// TODO: skip validation for now because of the following issue:
		// the bicep-types-az parses float as integer type and it should be fixed: https://github.com/Azure/bicep-types-az/issues/1404
		return nil
	case int64:
		v = int(input)
	case int32:
		v = int(input)
	case int:
		v = input
	default:
		return []error{ErrorMismatch(path, "integer", fmt.Sprintf("%T", body))}
	}
	if t.MinValue != nil && float64(v) < *t.MinValue {
		return []error{ErrorCommon(path, fmt.Sprintf("value is less than %v", *t.MinValue))}
	}
	if t.MaxValue != nil && float64(v) > *t.MaxValue {
		return []error{ErrorCommon(path, fmt.Sprintf("value is greater than %v", *t.MaxValue))}
	}
	return nil
}

func (t *NumberType) FilterReadOnlyFields(i interface{}) interface{} {
	return i
}

func (t *NumberType) FilterConfigurableFields(i interface{}) interface{} {
	return i
}

func (t *NumberType) AsTypeBase() *TypeBase {
	typeBase := TypeBase(t)
	return &typeBase
}
