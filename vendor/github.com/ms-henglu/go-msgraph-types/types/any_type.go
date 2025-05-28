package types

var _ TypeBase = &AnyType{}

type AnyType struct {
	Type        string `json:"$type"`
	Description string `json:"description"`
}

func (t *AnyType) Validate(i interface{}, s string) []error {
	return nil
}

func (t *AnyType) FilterReadOnlyFields(i interface{}) interface{} {
	return i
}

func (t *AnyType) FilterConfigurableFields(i interface{}) interface{} {
	return i
}

func (t *AnyType) AsTypeBase() *TypeBase {
	typeBase := TypeBase(t)
	return &typeBase
}
