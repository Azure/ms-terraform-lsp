resource "azapi_resource" "sg_resource" {
  type = "Microsoft.Relationships/serviceGroup@2023-09-01-preview"
  name = "sg0"
  parent_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-rg"
  body = {
    properties = {
      displayName = "example-service-group"
      description = "An example service group"
    }
  }
}

resource "azapi_resource" "sgm_resource1" {
  type = "Microsoft.Relationships/serviceGroupMember@2023-09-01-preview"
  name = "rel0"
  parent_id = azapi_resource.storage_resource.id
  body = {
    properties = {
      targetId = azapi_resource.sg_resource.id
    }
  }
  # schema_validation_enabled = false
}

resource "azapi_resource" "sgm_resource2" {
  type = "Microsoft.Relationships/serviceGroupMember@2023-09-01-preview"
  name = "rel1"
  parent_id = azapi_resource.storage_resource.id
  body = {
    properties = {
      targetId = azapi_resource.sg_resource.id
      metadata = {
        sourceType = ""
        targetType = ""
      }
      originInformation = {
        relationshipOriginType = ""
      }
      sourceId = azapi_resource.storage_resource.id
    }
  }
}