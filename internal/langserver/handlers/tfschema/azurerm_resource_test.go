package tfschema_test

import (
	"strings"
	"testing"

	"github.com/Azure/azurerm-lsp/internal/langserver/handlers/tfschema"
)

func TestAzureRMResource_Match(t *testing.T) {
	r := &tfschema.AzureRMResource{}

	testcases := []struct {
		name     string
		expected bool
	}{
		{
			name:     "resource.azurerm_resource",
			expected: true,
		},
		{
			name:     "resource.other_resource",
			expected: false,
		},
		{
			name:     "resource.azurerm_resource.other",
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

func TestAzureRMResource_ResourceDocumentation(t *testing.T) {
	r := &tfschema.AzureRMResource{}

	testcases := []struct {
		resourceType string
		expected     string
	}{
		{
			resourceType: "resource.azurerm_resource_group",
			expected:     `[📖 Documentation](<https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group>)`,
		},
		{
			resourceType: "resource.azurerm_virtual_network",
			expected:     `[📖 Documentation](<https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network>)`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.resourceType, func(t *testing.T) {
			if got := r.ResourceDocumentation(tc.resourceType); !strings.Contains(got, tc.expected) {
				t.Errorf("ResourceDocumentation(%s) = %s; want %s", tc.resourceType, got, tc.expected)
			}
		})
	}
}

func TestAzureRMResource_ListProperties(t *testing.T) {
	r := &tfschema.AzureRMResource{}

	testcases := []struct {
		blockPath string
		expected  []string
	}{
		{
			blockPath: "resource.azurerm_resource_group.",
			expected:  []string{"location", "name", "managed_by", "tags"},
		},

		{
			blockPath: "resource.azurerm_data_factory.identity",
			expected:  []string{"type", "identity_ids"},
		},

		{
			blockPath: "resource.azurerm_virtual_network.not_found",
			expected:  []string{},
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
				if prop.Name != tc.expected[i] {
					t.Errorf("ListProperties(%s)[%d] = %s; want %s", tc.blockPath, i, prop.Name, tc.expected[i])
				}
			}
		})
	}
}

func TestAzureRMResource_GetProperty(t *testing.T) {
	r := &tfschema.AzureRMResource{}

	testcases := []struct {
		propertyPath string
		expected     tfschema.Property
	}{
		{
			propertyPath: "resource.azurerm_resource_group.location",
			expected: tfschema.Property{
				Name:     "location",
				Modifier: "Required",
				Type:     "string",
			},
		},
		{
			propertyPath: "resource.azurerm_data_factory.identity.identity_ids",
			expected: tfschema.Property{
				Name:     "identity_ids",
				Modifier: "Optional",
				Type:     "list",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.propertyPath, func(t *testing.T) {
			prop := r.GetProperty(tc.propertyPath)
			if prop == nil {
				t.Errorf("GetProperty(%s) = nil; want %v", tc.propertyPath, tc.expected)
				return
			}
			if prop.Name != tc.expected.Name {
				t.Errorf("GetProperty(%s).Name = %s; want %s", tc.propertyPath, prop.Name, tc.expected.Name)
			}
			if prop.Modifier != tc.expected.Modifier {
				t.Errorf("GetProperty(%s).Modifier = %s; want %s", tc.propertyPath, prop.Modifier, tc.expected.Modifier)
			}
			if prop.Type != tc.expected.Type {
				t.Errorf("GetProperty(%s).Type = %s; want %s", tc.propertyPath, prop.Type, tc.expected.Type)
			}
		})
	}
}
