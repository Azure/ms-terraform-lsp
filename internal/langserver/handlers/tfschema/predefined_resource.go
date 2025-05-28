package tfschema

var _ Resource = &PredefinedResource{}

type PredefinedResource struct {
	Name       string
	Properties []Property
}

func (r *PredefinedResource) Match(name string) bool {
	if r == nil {
		return false
	}
	return r.Name == name
}

func (r *PredefinedResource) GetProperty(propertyName string) *Property {
	if r == nil {
		return nil
	}
	for _, prop := range r.Properties {
		if prop.Name == propertyName {
			return &prop
		}
	}
	return nil
}
