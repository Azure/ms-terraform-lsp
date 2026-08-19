resource "azapi_resource_action" "test" {
  type = "Microsoft.Network/networkManagers@2024-01-01-preview"
  action = "listDeploymentStatus"
  body = jsonencode({

  })
}
