module "avm" {
  source = "Azure/avm-res-storage-storageaccount/azurerm"

  name                = "example-account"
  azure_files_authentication =
}