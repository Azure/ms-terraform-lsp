package processors

import (
	"embed"
	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
)

//go:embed .tools/generate-provider-schema/*
var runScript embed.FS

func ProcessSchema(providerPath, gitBranch string) (*schema.ProviderSchema, error) {

	return &schema.ProviderSchemaInfo, nil
}
