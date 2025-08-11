package schema

const (
	TerraformDocsURL  = "https://registry.terraform.io/providers/hashicorp/azurerm/%[1]s/docs/%[2]s/%[3]s"
	AVMDocsURL        = "https://registry.terraform.io/modules/Azure/%[1]s/azurerm/%[2]s"
	DefaultDocVersion = "latest"
	Resources         = "resources"
	DataSources       = "data-sources"
	AzureRMPrefix     = "azurerm_"
	AVMPrefix         = "Azure/avm-"
	ResourcesPrefix   = "resource"
	DataSourcesPrefix = "data"
)

const (
	InputDataSourcePrefix = "datasource#"
)

const (
	AVMGitHubIssuesURL          = "https://github.com/Azure/terraform-azurerm-%[1]s/issues"
	AVMAttributeDocURL          = "https://github.com/Azure/terraform-azurerm-%[1]s/blob/main/README.md#-%[2]s"
	AVMGitHubAttributeIssuesURL = "https://github.com/Azure/terraform-azurerm-%[1]s/issues?q=is:issue %[2]s"
	AVMNewGitHubIssuesURL       = "https://github.com/Azure/terraform-azurerm-%[1]s/issues/new?template=avm_module_issue.yml"
	GitHubIssuesURL             = "https://github.com/hashicorp/terraform-provider-azurerm/issues?q=is:issue %[1]s"
	NewGitHubIssuesURL          = "https://github.com/hashicorp/terraform-provider-azurerm/issues/new?template=Bug_Report.yml&title=%[1]s "
)
