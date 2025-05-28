package types

import (
	"fmt"
	"strconv"
)

var _ TypeBase = &ArrayType{}

type ArrayType struct {
	Type      string         `json:"$type"`
	ItemType  *TypeReference `json:"itemType"`
	MinLength *uint64        `json:"minLength"`
	MaxLength *uint64        `json:"maxLength"`
}

func (t *ArrayType) Validate(body interface{}, path string) []error {
	if t == nil || body == nil {
		return []error{}
	}
	errors := make([]error, 0)
	var itemType TypeBase
	if t.ItemType != nil {
		itemType = t.ItemType.Type
	}
	// check body type
	bodyArray, ok := body.([]interface{})
	if !ok {
		errors = append(errors, ErrorMismatch(path, "array", fmt.Sprintf("%T", body)))
		return errors
	}

	// check the length
	if t.MinLength != nil && uint64(len(bodyArray)) < *t.MinLength {
		errors = append(errors, ErrorCommon(path, fmt.Sprintf("array length is less than %d", *t.MinLength)))
	}

	if t.MaxLength != nil && uint64(len(bodyArray)) > *t.MaxLength {
		errors = append(errors, ErrorCommon(path, fmt.Sprintf("array length is greater than %d", *t.MaxLength)))
	}

	for index, value := range bodyArray {
		if itemType != nil {
			errors = append(errors, itemType.Validate(value, path+"."+strconv.Itoa(index))...)
		}
	}
	return errors
}

func (t *ArrayType) FilterReadOnlyFields(i interface{}) interface{} {
	if t == nil || i == nil {
		return nil
	}
	if t.ItemType == nil {
		return i
	}

	itemType := t.ItemType.Type
	// check body type
	bodyArray, ok := i.([]interface{})
	if !ok {
		return nil
	}

	res := make([]interface{}, 0)
	for _, value := range bodyArray {
		res = append(res, itemType.FilterReadOnlyFields(value))
	}
	return res
}

func (t *ArrayType) FilterConfigurableFields(i interface{}) interface{} {
	if t == nil || i == nil {
		return nil
	}

	if t.ItemType == nil {
		return i
	}

	itemType := t.ItemType.Type
	// check body type
	bodyArray, ok := i.([]interface{})
	if !ok {
		return nil
	}

	res := make([]interface{}, 0)
	for _, value := range bodyArray {
		res = append(res, itemType.FilterConfigurableFields(value))
	}
	return res
}

func (t *ArrayType) AsTypeBase() *TypeBase {
	typeBase := TypeBase(t)
	return &typeBase
}
