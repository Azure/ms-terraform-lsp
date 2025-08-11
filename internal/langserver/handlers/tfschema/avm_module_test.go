package tfschema_test

import (
	"strings"
	"testing"

	"github.com/Azure/ms-terraform-lsp/internal/langserver/handlers/tfschema"
)

func TestAVMModule_Match(t *testing.T) {
	r := &tfschema.AVMModule{}

	testcases := []struct {
		name     string
		expected bool
	}{
		{
			name:     "module.avm-res-compute-virtualmachine",
			expected: true,
		},
		{
			name:     "resource.azurerm_resource",
			expected: false,
		},
		{
			name:     "module.other_module",
			expected: true,
		},
		{
			name:     "azurerm_resource.other",
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

func TestAVMModule_ResourceDocumentation(t *testing.T) {
	r := &tfschema.AVMModule{}

	testcases := []struct {
		resourceType string
		expected     string
	}{
		{
			resourceType: "module.Azure/avm-res-compute-virtualmachine/azurerm",
			expected:     `[📖 Documentation](<https://registry.terraform.io/modules/Azure/avm-res-compute-virtualmachine/azurerm/latest>)`,
		},
		{
			resourceType: "module.Azure/avm-res-network-virtualnetwork/azurerm",
			expected:     `[📖 Documentation](<https://registry.terraform.io/modules/Azure/avm-res-network-virtualnetwork/azurerm/latest>)`,
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

func TestAVMModule_ListProperties(t *testing.T) {
	r := &tfschema.AVMModule{}

	testcases := []struct {
		blockPath string
		expected  []string
	}{
		{
			blockPath: "module.Azure/avm-res-compute-virtualmachine/azurerm",
			expected:  []string{"account_credentials", "additional_unattend_contents", "boot_diagnostics", "data_disk_managed_disks", "diagnostic_settings", "enable_telemetry", "extensions", "gallery_applications", "location", "managed_identities", "name", "network_interfaces", "os_disk", "os_type", "public_ip_configuration_details", "resource_group_name", "role_assignments", "role_assignments_system_managed_identity", "secrets", "sku_size", "winrm_listeners", "allow_extension_operations", "availability_set_resource_id", "azure_backup_configurations", "boot_diagnostics_storage_account_uri", "bypass_platform_safety_checks_on_user_schedule_enabled", "capacity_reservation_group_resource_id", "computer_name", "custom_data", "dedicated_host_group_resource_id", "dedicated_host_resource_id", "disk_controller_type", "edge_zone", "enable_automatic_updates", "encryption_at_host_enabled", "eviction_policy", "extensions_time_budget", "hotpatching_enabled", "license_type", "lock", "maintenance_configuration_resource_ids", "max_bid_price", "patch_assessment_mode", "patch_mode", "plan", "platform_fault_domain", "priority", "provision_vm_agent", "proximity_placement_group_resource_id", "reboot_setting", "run_commands", "run_commands_secrets", "secure_boot_enabled", "shutdown_schedules", "source_image_reference", "source_image_resource_id", "tags", "termination_notification", "timeouts", "timezone", "user_data", "virtual_machine_scale_set_resource_id", "vm_additional_capabilities", "vtpm_enabled", "zone"},
		},
		{
			blockPath: "module.Azure/avm-res-compute-virtualmachine/azurerm.os_disk",
			expected:  []string{"caching", "storage_account_type", "diff_disk_settings", "disk_encryption_set_id", "disk_size_gb", "name", "secure_vm_disk_encryption_set_id", "security_encryption_type", "write_accelerator_enabled"},
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

func TestAVMModule_GetProperty(t *testing.T) {
	r := &tfschema.AVMModule{}

	testcases := []struct {
		propertyPath string
		expected     tfschema.Property
	}{
		{
			propertyPath: "module.Azure/avm-res-compute-virtualmachine/azurerm.location",
			expected: tfschema.Property{
				Name:     "location",
				Modifier: "Required",
				Type:     "string",
			},
		},
		{
			propertyPath: "module.Azure/avm-res-compute-virtualmachine/azurerm.os_disk.caching",
			expected: tfschema.Property{
				Name:     "caching",
				Modifier: "Required",
				Type:     "string",
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
			if !strings.HasPrefix(prop.Type, tc.expected.Type) {
				t.Errorf("GetProperty(%s).Type = %s; want %s", tc.propertyPath, prop.Type, tc.expected.Type)
			}
		})
	}
}
