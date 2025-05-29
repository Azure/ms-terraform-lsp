package tfschema

import lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"

var Resources []Resource

func init() {
	Resources = make([]Resource, 0)

	Resources = append(Resources,
		&AzureRMResource{},
		&MSGraphResource{
			Name: "resource.msgraph_resource",
			Properties: []Property{
				{
					Name:                  "url",
					Modifier:              "Required",
					Type:                  "string",
					Description:           "MS Graph API URL. Collection URL which should be used to create the msgraph resource. For example, `/users`.",
					CompletionNewText:     `url = "$0"`,
					GenericCandidatesFunc: urlCandidates,
				},

				{
					Name:                "api_version",
					Modifier:            "Optional",
					Type:                "string",
					Description:         "MS Graph API version. Allowed values are `v1.0` and `beta`. Defaults to `v1.0`.",
					CompletionNewText:   `api_version = "$0"`,
					ValueCandidatesFunc: apiVersionCandidates,
				},

				{
					Name:                  "body",
					Modifier:              "Optional",
					Type:                  "dynamic",
					Description:           "An HCL object that contains the request body used to create and update msgraph resource.",
					CompletionNewText:     `body = $0`,
					ValueCandidatesFunc:   FixedValueCandidatesFunc([]lsp.CompletionItem{dynamicPlaceholderCandidate()}),
					GenericCandidatesFunc: bodyCandidates,
					CustomizedHoverFunc:   msgraphBodyHover,
				},

				{
					Name:              "response_export_values",
					Modifier:          "Optional",
					Type:              "map<string, string>",
					Description:       "The attribute can accept a map of path that needs to be exported from response body.",
					CompletionNewText: `response_export_values = $0`,
				},

				{
					Name:              "headers",
					Modifier:          "Optional",
					Type:              "map<string, string>",
					Description:       "A mapping of headers which should be sent with the request.",
					CompletionNewText: `headers = $0`,
				},

				{
					Name:              "query_parameters",
					Modifier:          "Optional",
					Type:              "map<string, list<string>>",
					Description:       "A mapping of query parameters which should be sent with the request.",
					CompletionNewText: `query_parameters = $0`,
				},
			},
		},
		&MSGraphResource{
			Name: "data.msgraph_resource",
			Properties: []Property{
				{
					Name:                  "url",
					Modifier:              "Required",
					Type:                  "string",
					Description:           "MS Graph API URL. Collection URL which should be used to create the msgraph resource. For example, `/users`.",
					CompletionNewText:     `url = "$0"`,
					GenericCandidatesFunc: urlCandidates,
				},

				{
					Name:                "api_version",
					Modifier:            "Optional",
					Type:                "string",
					Description:         "MS Graph API version. Allowed values are `v1.0` and `beta`. Defaults to `v1.0`.",
					CompletionNewText:   `api_version = "$0"`,
					ValueCandidatesFunc: apiVersionCandidates,
				},

				{
					Name:              "response_export_values",
					Modifier:          "Optional",
					Type:              "map<string, string>",
					Description:       "The attribute can accept a map of path that needs to be exported from response body.",
					CompletionNewText: `response_export_values = $0`,
				},

				{
					Name:              "headers",
					Modifier:          "Optional",
					Type:              "map<string, string>",
					Description:       "A mapping of headers which should be sent with the request.",
					CompletionNewText: `headers = $0`,
				},

				{
					Name:              "query_parameters",
					Modifier:          "Optional",
					Type:              "map<string, list<string>>",
					Description:       "A mapping of query parameters which should be sent with the request.",
					CompletionNewText: `query_parameters = $0`,
				},
			},
		},
	)
}
