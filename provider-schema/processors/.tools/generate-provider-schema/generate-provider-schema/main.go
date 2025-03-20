package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/provider"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	tfyschema "github.com/magodo/tfadd/schema/legacy"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	resources := map[string]*schema.Resource{}
	dataSources := map[string]*schema.Resource{}

	for _, service := range provider.SupportedTypedServices() {
		for _, rs := range service.Resources() {
			wrapper := sdk.NewResourceWrapper(rs)
			rsWrapper, err := wrapper.Resource()
			if err != nil {
				return fmt.Errorf("wrapping Resource %q: %+v", rs.ResourceType(), err)
			}
			resources[rs.ResourceType()] = rsWrapper
		}

		for _, ds := range service.DataSources() {
			wrapper := sdk.NewDataSourceWrapper(ds)
			dsWrapper, err := wrapper.DataSource()
			if err != nil {
				return fmt.Errorf("wrapping DataSource %q: %+v", ds.ResourceType(), err)
			}
			dataSources[ds.ResourceType()] = dsWrapper
		}
	}

	for _, service := range provider.SupportedUntypedServices() {
		for name, rs := range service.SupportedResources() {
			resources[name] = rs
		}

		for name, ds := range service.SupportedDataSources() {
			dataSources[name] = ds
		}
	}

	schemas := map[string]*tfyschema.Schema{}
	for name, res := range resources {
		schemas[name] = &tfyschema.Schema{Block: tfyschema.FromProviderSchemaMap(res.Schema)}
	}
	for name, ds := range dataSources {
		schemas["datasource#"+name] = &tfyschema.Schema{Block: tfyschema.FromProviderSchemaMap(ds.Schema)}
	}

	b, err := json.Marshal(tfyschema.ProviderSchema{ResourceSchemas: schemas})
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
