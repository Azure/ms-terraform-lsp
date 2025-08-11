// ---------------------------------------------------------------------------------
// Core Configuration
// These variables define the base configuration and location for the deployment.
// ---------------------------------------------------------------------------------
variable "subscription_id" {
  description = "The Azure subscription ID."
  default = ""

}
variable "base_name" {
  description = "This is the base name for each Azure resource name."
}

variable "location" {
  description = "The resource group location."

}

variable "tags" {
  description = "Map of tags to add to resources."
  type        = map(string)
  default     = {}
}

// ---------------------------------------------------------------------------------
// Resource Management
// These variables control the use of existing resources.
// ---------------------------------------------------------------------------------

variable "use_existing_rg" {
  description = "Flag to determine if an existing resource group should be used."
  type        = bool
  default     = false
}

variable "existing_rg_name" {
  description = "Name of the existing resource group to use."
  type        = string
  default     = ""
}

variable "existing_vnet_id" {
  description = "The ID of an existing virtual network to use. If not defined, a new one will be used."
  type        = string
  default     = null
}

variable "existing_subnet_id" {
  description = "The ID of an existing subnet to use. If not defined, a new one will be used."
  type        = string
  default     = null
}

// ---------------------------------------------------------------------------------
// Networking
// Configuration options related to network settings.
// ---------------------------------------------------------------------------------

variable "deploy_network" {
  description = "Flag to deploy network resources."
  type        = bool
  default     = true
}

variable "network" {
  description = "Network configuration."
  type = object({
    base_name                       = string

    development_environment         = bool
    vnet_address_prefix             = string
    app_gateway_subnet_prefix       = string
    private_endpoints_subnet_prefix = string
    agents_subnet_prefix            = string
    bastion_subnet_prefix           = string
    jumpbox_subnet_prefix           = string
    training_subnet_prefix          = string
    scoring_subnet_prefix           = string
    app_services_subnet_prefix      = string

  })
  default = {
    base_name                       = "example"
    development_environment         = true
    vnet_address_prefix             = "10.0.0.0/16"
    app_gateway_subnet_prefix       = "10.0.1.0/24"
    private_endpoints_subnet_prefix = "10.0.2.0/27"
    agents_subnet_prefix            = "10.0.2.32/27"
    bastion_subnet_prefix           = "10.0.2.64/26"
    jumpbox_subnet_prefix           = "10.0.2.128/28"
    training_subnet_prefix          = "10.0.3.0/24"
    scoring_subnet_prefix           = "10.0.4.0/24"
    app_services_subnet_prefix      = "10.0.5.0/24"
  }
}

variable "jumpbox_config" {
  default = {
    log_workspace_name           = ""
    jump_box_admin_name          = "vmadmin"
    jump_box_admin_password      = ""
    vm_size                      = "Standard_DS1_v2"
    image_publisher              = "MicrosoftWindowsServer"
    image_offer                  = "WindowsServer"
    image_sku                    = "2019-Datacenter"
    image_version                = "latest"
    os_disk_caching              = "ReadWrite"
    os_disk_storage_account_type = "Standard_LRS"
  }

}

// ---------------------------------------------------------------------------------
// Role and Access Management
// Variables defining role templates and group assignments.
// ---------------------------------------------------------------------------------

variable "role_templates" {
  description = "Templates for role assignments."
  type = map(list(object({
    role_name = string
    scope     = string
  })))
  default = {
    infra_admin = [
      { role_name = "contributor", scope = "resource_group_id" },
      { role_name = "azure_ai_administrator", scope = "resource_group_id" },
      { role_name = "search_index_data_contributor", scope = "ai_search_service_id" },
      { role_name = "cognitive_services_openai_user", scope = "openai_embedding_id" },
      { role_name = "cognitive_services_openai_contributor", scope = "openai_chat_id" },
      { role_name = "search_service_contributor", scope = "ai_search_service_id" },
      { role_name = "storage_blob_data_contributor", scope = "storage_account_id" },
      { role_name = "storage_file_data_privileged_contributor", scope = "storage_account_id" }
    ]
    ai_admin = [
      { role_name = "owner", scope = "ai_hub_id" },
      { role_name = "azure_ai_administrator", scope = "resource_group_id" },
      { role_name = "search_index_data_contributor", scope = "ai_search_service_id" },
      { role_name = "search_service_contributor", scope = "ai_search_service_id" },
      { role_name = "cognitive_services_openai_contributor", scope = "openai_chat_id" },
      { role_name = "cognitive_services_openai_user", scope = "openai_embedding_id" },
      { role_name = "storage_blob_data_contributor", scope = "storage_account_id" },
      { role_name = "storage_file_data_privileged_contributor", scope = "storage_account_id" }
    ]
  }
}

variable "group_assignments" {
  description = "Assignments for groups."
  default = {}
}


// ---------------------------------------------------------------------------------
// Deployment Configurations
// These settings facilitate the deployment process and environment specifics.
// ---------------------------------------------------------------------------------

variable "development_environment" {
  description = "Optional flag to deploy a development environment."
  type        = bool
}

variable "deployment_config" {
  description = "Configuration to choose which layers to deploy."
  type = object({
    deploy_network  = bool
    deploy_services = bool
    deploy_core     = bool
    deploy_identity = bool
    deploy_shared   = bool
  })
  default = {
    deploy_network  = true
    deploy_services = true
    deploy_core     = true
    deploy_identity = true
    deploy_shared   = true
  }

  validation {
    condition     = !var.deployment_config.deploy_identity || (var.deployment_config.deploy_services && var.deployment_config.deploy_core)
    error_message = "The identity module depends on both services and core modules. Deploy those before deploying identity."
  }

  validation {
    condition     = !var.deployment_config.deploy_shared || var.deployment_config.deploy_services
    error_message = "The shared module requires the services module. Deploy services before shared."
  }
}

// Additional configuration for module extensions
variable "extra_shared_private_links" {
  description = "Additional shared private links to configure."
  type = list(object({
    groupId               = string
    status                = string
    provisioningState     = string
    requestMessage        = string
    privateLinkResourceId = string
  }))
  default = []
}

variable "extra_ai_hub_outbound_rules" {
  description = "Additional AI Hub outbound rules to configure."
  type = map(object({
    type = string
    destination = object({
      serviceResourceId = string
      subresourceTarget = string
      sparkEnabled      = bool
      sparkStatus       = string
    })
  }))
  default = {}
}

// ---------------------------------------------------------------------------------
// Service Configurations
// Specifies the expected properties for the specific services, including search and AI services.
// ---------------------------------------------------------------------------------

variable "search_config" {
  description = "Configuration for the search service."
  type = object({
    private_dns_zone_ids  = list(string)
    tags                  = map(string)
    sku_name              = string
    disable_local_auth    = bool
    hosting_mode          = string
    public_network_access = string
    partition_count       = number
    replica_count         = number
    semantic_search       = string
    search_identity_provider = object({
      type = string
    })
    deploy_shared_private_link = bool
    deploy_private_dns_zones   = bool
  })
  default = {
    private_dns_zone_ids  = []
    tags                  = {}
    sku_name              = "standard"
    disable_local_auth    = true
    hosting_mode          = "default"
    public_network_access = "disabled"
    partition_count       = 1
    replica_count         = 1
    semantic_search       = "disabled"
    search_identity_provider = {
      type = "None"
    }
    deploy_shared_private_link = false
    deploy_private_dns_zones   = true
  }
}

variable "aiservice_config" {
  description = "Configuration for the AI service."
  type = object({
    private_dns_zone_ids = list(string)
    aiServiceSkuName     = string

    disableLocalAuth         = bool
    deploy_private_dns_zones = bool
  })
  default = {
    private_dns_zone_ids = []
    aiServiceSkuName     = "S0"

    disableLocalAuth         = false
    deploy_private_dns_zones = true
  }
}

variable "core_config" {
  description = "Configuration for ai-foundry-core module."
  type = object({
    acr = object({
      private_dns_zone_ids   = list(string)
      deploy_acr_private_dns = bool
    })
    storage = object({
      private_dns_zone_ids       = list(string)
      deploy_storage_private_dns = bool
    })
    key_vault = object({
      private_dns_zone_ids       = list(string)
      deploy_storage_private_dns = bool
    })
    ai_hub = object({
      private_dns_zone_ids = list(string)
      tags                 = map(string)
      deploy_private_dns   = bool
      description          = string
    })
  })
  default = {
    acr = {
      private_dns_zone_ids   = []
      deploy_acr_private_dns = true
    }
    storage = {
      private_dns_zone_ids       = []
      deploy_storage_private_dns = true
    }
    key_vault = {
      private_dns_zone_ids       = []
      deploy_storage_private_dns = true
    }
    ai_hub = {
      private_dns_zone_ids = []
      tags                 = {}
      deploy_private_dns   = true
      description          = "AI Hub"
    }
  }
}


