package tfschema_test

import (
	"testing"

	"github.com/Azure/ms-terraform-lsp/internal/langserver/handlers/tfschema"
)

func TestMSGraphResource_Match(t *testing.T) {
	r := &tfschema.MSGraphResource{
		Name: "resource.msgraph_resource",
	}

	testcases := []struct {
		name     string
		expected bool
	}{
		{
			name:     "resource.msgraph_resource",
			expected: true,
		},
		{
			name:     "resource.other_resource",
			expected: false,
		},
		{
			name:     "resource.msgraph_resource.other",
			expected: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := r.Match(tc.name); got != tc.expected {
				t.Errorf("Match(%s) = %v; want %v", tc.name, got, tc.expected)
			}
		})
	}
}

func TestMSGraphResource_ResourceDocumentation(t *testing.T) {
	r := &tfschema.MSGraphResource{
		Name: "resource.msgraph_resource",
	}

	testcases := []struct {
		resourceType string
		expected     string
	}{
		{
			resourceType: "applications@v1.0",
			expected:     "Url: 'applications'  \nSummary: Create application  \nDescription: Create a new application object.  \n\n[Find more info here](https://learn.microsoft.com/graph/api/application-post-applications?view=graph-rest-1.0)",
		},
		{
			resourceType: "users@v1.0",
			expected:     "Url: 'users'  \nSummary: Create user  \nDescription: Create a new user object.  \n\n[Find more info here](https://learn.microsoft.com/graph/api/intune-onboarding-user-create?view=graph-rest-1.0)",
		},
		{
			resourceType: "invalid_resource",
			expected:     "",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.resourceType, func(t *testing.T) {
			doc := r.ResourceDocumentation(tc.resourceType)
			if doc != tc.expected {
				t.Errorf("ResourceDocumentation(%s) = %q; want %q", tc.resourceType, doc, tc.expected)
			}
		})
	}
}

func TestMSGraphResource_ListProperties(t *testing.T) {
	r := &tfschema.MSGraphResource{
		Name: "resource.msgraph_resource",
		Properties: []tfschema.Property{
			{Name: "property1", Type: "string"},
			{Name: "property2", Type: "int"},
			{Name: "nestedProperty", NestedProperties: []tfschema.Property{
				{Name: "nestedProperty1", Type: "bool"},
			}},
		},
	}

	testcases := []struct {
		blockPath string
		expected  []tfschema.Property
	}{
		{
			blockPath: "resource.msgraph_resource",
			expected:  nil,
		},
		{
			blockPath: "resource.msgraph_resource.",
			expected:  r.Properties,
		},
		{
			blockPath: "resource.msgraph_resource.nestedProperty",
			expected:  r.GetProperty("resource.msgraph_resource.nestedProperty").NestedProperties,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.blockPath, func(t *testing.T) {
			props := r.ListProperties(tc.blockPath)
			if len(props) != len(tc.expected) {
				t.Errorf("ListProperties(%s) = %d properties; want %d", tc.blockPath, len(props), len(tc.expected))
				return
			}
			for i, prop := range props {
				if prop.Name != tc.expected[i].Name || prop.Type != tc.expected[i].Type {
					t.Errorf("ListProperties(%s)[%d] = %v; want %v", tc.blockPath, i, prop, tc.expected[i])
				}
			}
		})
	}
}

func TestMSGraphResource_GetProperty(t *testing.T) {
	r := &tfschema.MSGraphResource{
		Name: "resource.msgraph_resource",
		Properties: []tfschema.Property{
			{Name: "property1", Type: "string"},
			{Name: "property2", Type: "int"},
			{Name: "nestedProperty", NestedProperties: []tfschema.Property{
				{Name: "nestedProperty1", Type: "bool"},
			}},
		},
	}

	testcases := []struct {
		propertyPath string
		expected     *tfschema.Property
	}{
		{
			propertyPath: "resource.msgraph_resource.property1",
			expected:     &tfschema.Property{Name: "property1", Type: "string"},
		},
		{
			propertyPath: "resource.msgraph_resource.property2",
			expected:     &tfschema.Property{Name: "property2", Type: "int"},
		},
		{
			propertyPath: "resource.msgraph_resource.nestedProperty.nestedProperty1",
			expected:     &tfschema.Property{Name: "nestedProperty1", Type: "bool"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.propertyPath, func(t *testing.T) {
			prop := r.GetProperty(tc.propertyPath)
			if prop == nil && tc.expected != nil {
				t.Errorf("GetProperty(%s) = nil; want %v", tc.propertyPath, tc.expected)
				return
			}
			if prop != nil && (prop.Name != tc.expected.Name || prop.Type != tc.expected.Type) {
				t.Errorf("GetProperty(%s) = %v; want %v", tc.propertyPath, prop, tc.expected)
			}
		})
	}
}
