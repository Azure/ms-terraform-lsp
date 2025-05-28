package schema

import "github.com/ms-henglu/go-msgraph-types/types"

func GetRequiredPropertySet(typeBase *types.TypeBase) []PropertySet {
	if typeBase == nil {
		return nil
	}
	switch t := (*typeBase).(type) {
	case *types.ObjectType:
		requiredProps := make(map[string]Property)
		for propName, prop := range t.Properties {
			if propName == "@odata.type" {
				continue
			}
			if prop.IsRequired() {
				if value := PropertyFromObjectProperty(propName, prop); value != nil {
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
			return GetRequiredPropertySet(&t.Body.Type)
		}
	case *types.UnionType:
		res := make([]PropertySet, 0)
		for _, element := range t.Elements {
			if element.Type == nil {
				continue
			}
			res = append(res, GetRequiredPropertySet(&element.Type)...)
		}
		return res
	}
	return nil
}
