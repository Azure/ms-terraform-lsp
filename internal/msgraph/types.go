package msgraph

import "github.com/ms-henglu/go-msgraph-types/types"

var SchemaLoader *types.MSGraphSchemaLoader

func init() {
	SchemaLoader = types.DefaultMSGraphSchemaLoader()
}
