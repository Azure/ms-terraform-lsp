module "avm" {
  source = "Azure/avm-res-servicebus-namespace/azurerm"

  name                = "example"
  minimum_tls_version = "1.2"
}