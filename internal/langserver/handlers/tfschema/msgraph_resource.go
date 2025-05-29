package tfschema

import (
	"fmt"
	"strings"

	"github.com/Azure/ms-terraform-lsp/internal/msgraph"
)

var _ Resource = &MSGraphResource{}

type MSGraphResource struct {
	Name       string
	Properties []Property
}

func (r *MSGraphResource) ResourceDocumentation(resourceType string) string {
	parts := strings.Split(resourceType, "@")
	if len(parts) != 2 {
		return ""
	}
	apiVersion := parts[1]
	urlValue := parts[0]
	resourceDef := msgraph.SchemaLoader.GetResourceDefinition(apiVersion, urlValue)
	doc := fmt.Sprintf("Url: '%s'  \nSummary: %s  \nDescription: %s  \n", resourceDef.Url, resourceDef.Name, resourceDef.Description)
	if resourceDef.ExternalDocs != nil {
		doc = fmt.Sprintf("%s\n[%s](%s)", doc, resourceDef.ExternalDocs.Description, resourceDef.ExternalDocs.Url)
	}
	return doc
}

func (r *MSGraphResource) ListProperties(blockPath string) []Property {
	p := r.GetProperty(blockPath)
	if p == nil {
		return nil
	}
	return p.NestedProperties
}

func (r *MSGraphResource) Match(name string) bool {
	if r == nil {
		return false
	}
	return r.Name == name
}

func (r *MSGraphResource) GetProperty(propertyPath string) *Property {
	if r == nil {
		return nil
	}
	parts := strings.Split(propertyPath, ".")
	if len(parts) <= 2 {
		return nil
	}

	p := Property{
		NestedProperties: r.Properties,
	}
	if parts[2] == "" {
		return &p
	}
	for index := 2; index < len(parts); index++ {
		found := false
		for _, prop := range p.NestedProperties {
			if prop.Name == parts[index] {
				p = prop
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return &p
}
