package azure_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/ms-terraform-lsp/internal/azure"
	"github.com/Azure/ms-terraform-lsp/internal/utils"
)

func Test_BodyValidation(t *testing.T) {
	testData := []struct {
		Id         string
		ApiVersion string
		Body       string
		Error      bool
	}{
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.MachineLearningServices/workspaces/workspace1/computes/compute1",
			ApiVersion: "2021-07-01",
			Body: `
{
    "location": "eastus",
    "properties": {
        "properties": {
            "state": "Running"
        }
    }
}
`,
			Error: true, // properties.computeType is required
		},
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.MachineLearningServices/workspaces/workspace1/computes/compute1",
			ApiVersion: "2021-07-01",
			Body: `
{
    "location": "eastus",
    "properties": {
        "computeType": "ComputeInstance1",
        "properties": {
            "state": "Running"
        }
    }
}
`,
			Error: true, // invalid properties.computeType
		},
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.MachineLearningServices/workspaces/workspace1/computes/compute1",
			ApiVersion: "2021-07-01",
			Body: `
{
    "location": "eastus",
    "properties": {
        "computeType": "ComputeInstance",
        "properties": {
            "state": "Running"
        }
    }
}
`,
			Error: true, // properties.properties.state is read only
		},
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.MachineLearningServices/workspaces/workspace1/computes/compute1",
			ApiVersion: "2021-07-01",
			Body: `
{
    "location": "eastus",
    "properties": {
        "computeType": "ComputeInstance",
        "properties": {
        }
    }
}
`,
			Error: false,
		},
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acctestRG-211109150453866525/providers/Microsoft.ContainerRegistry/registries/acctest61311",
			ApiVersion: "2020-11-01-preview",
			Body: `
{
  "identity": {
    "type": "SystemAssigned, UserAssigned",
    "userAssignedIdentities": {
      "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acctestRG-211109152718418172/providers/Microsoft.ManagedIdentity/userAssignedIdentities/acctestb0i47": {}
    }
  },
  "location": "westeurope",
  "properties": {
    "adminUserEnabled": true
  },
  "sku": {
    "name": "Standard"
  },
  "tags": {
    "Key": "Value"
  }
}
`,
			Error: false,
		},
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroup/providers/Microsoft.AppPlatform/Spring/myservice/apps/myapp",
			ApiVersion: "2020-07-01",
			Body: `
{
    "location": "westeurope",
    "properties": {
        "activeDeploymentName": "mydeployment1",
        "fqdn": "myapp.mydomain.com",
        "httpsOnly": 1,
        "persistentDisk": {
            "mountPath": "/persistent",
            "sizeInGB": 2
        },
        "publi1c": true,
        "temporaryDisk": {
            "mountPath": "/temp",
            "sizeInGB": 2
        }
    }
}
`,
			Error: true,
		},
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroup/providers/Microsoft.Automation/automationAccounts/acctest3093",
			ApiVersion: "2021-06-22",
			Body: `
{
    "location": "westeurope",
    "name": "myAutomationAccount9",
    "properties": {
        "sku": {
            "name": "Free"
        }
    }
}
`,
			Error: false,
		},
	}

	for index, data := range testData {
		resourceType := utils.GetResourceType(data.Id)

		var body interface{}
		_ = json.Unmarshal([]byte(data.Body), &body)

		def, err := azure.GetResourceDefinition(resourceType, data.ApiVersion)
		if err != nil {
			t.Fatal(err)
		}

		if def != nil {
			errors := (*def).Validate(body, "")
			fmt.Printf("Running test for case %d, resource type: %s, api-version: %s\n", index, resourceType, data.ApiVersion)
			fmt.Println(errors)
			if (len(errors) > 0) != data.Error {
				t.Errorf("expect error: %t, got error: %t for id: %s, api-version: %s", data.Error, len(errors) > 0, data.Id, data.ApiVersion)
			}
		} else {
			t.Fatalf("failed to load resource definition for id: %s, api-version: %s", data.Id, data.ApiVersion)
		}
	}
}

func Test_WriteOnly(t *testing.T) {
	testData := []struct {
		Id         string
		ApiVersion string
		Input      string
		Output     string
	}{
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb",
			ApiVersion: "2021-03-01",
			Input: `
{
    "name": "mylb",
    "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb",
    "etag": "W/\"4cdb755a-4607-49eb-b52c-2f5d06a9a0d5\"",
    "type": "Microsoft.Network/loadBalancers",
    "location": "westus",
    "tags": {},
    "properties": {
        "provisioningState": "Succeeded",
        "resourceGuid": "6a319c45-812d-4a3f-bd9b-edfcd9017037",
        "frontendIPConfigurations": [
            {
                "name": "PublicIPAddress",
                "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb/frontendIPConfigurations/PublicIPAddress",
                "etag": "W/\"4cdb755a-4607-49eb-b52c-2f5d06a9a0d5\"",
                "type": "Microsoft.Network/loadBalancers/frontendIPConfigurations",
                "properties": {
                    "provisioningState": "Succeeded",
                    "privateIPAllocationMethod": "Dynamic",
                    "publicIPAddress": {
                        "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/myip"
                    },
                    "inboundNatRules": [
                        {
                            "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb/inboundNatRules/RDPAccess"
                        }
                    ]
                }
            }
        ],
        "backendAddressPools": [],
        "loadBalancingRules": [],
        "probes": [],
        "inboundNatRules": [
            {
                "name": "RDPAccess",
                "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb/inboundNatRules/RDPAccess",
                "etag": "W/\"4cdb755a-4607-49eb-b52c-2f5d06a9a0d5\"",
                "type": "Microsoft.Network/loadBalancers/inboundNatRules",
                "properties": {
                    "provisioningState": "Succeeded",
                    "frontendIPConfiguration": {
                        "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb/frontendIPConfigurations/PublicIPAddress"
                    },
                    "frontendPort": 3389,
                    "backendPort": 3389,
                    "enableFloatingIP": false,
                    "idleTimeoutInMinutes": 4,
                    "protocol": "Tcp",
                    "enableDestinationServiceEndpoint": false,
                    "enableTcpReset": false,
                    "allowBackendPortConflict": false
                }
            }
        ],
        "inboundNatPools": []
    },
    "sku": {
        "name": "Basic",
        "tier": "Regional"
    }
}
`,
			Output: `
{
    "name": "mylb",
    "location": "westus",
    "tags": {},
    "properties": {
        "frontendIPConfigurations": [
            {
                "name": "PublicIPAddress",
                "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb/frontendIPConfigurations/PublicIPAddress",
                "properties": {
                    "privateIPAllocationMethod": "Dynamic",
                    "publicIPAddress": {
                        "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/myip"
                    }
                }
            }
        ],
        "backendAddressPools": [],
        "loadBalancingRules": [],
        "probes": [],
        "inboundNatRules": [
            {
                "name": "RDPAccess",
                "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb/inboundNatRules/RDPAccess",
                "properties": {
                    "frontendIPConfiguration": {
                        "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/mylb/frontendIPConfigurations/PublicIPAddress"
                    },
                    "frontendPort": 3389,
                    "backendPort": 3389,
                    "enableFloatingIP": false,
                    "idleTimeoutInMinutes": 4,
                    "protocol": "Tcp",
                    "enableTcpReset": false
                }
            }
        ],
        "inboundNatPools": []
    },
    "sku": {
        "name": "Basic",
        "tier": "Regional"
    }
}
`,
		},
		{
			Id:         "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.MachineLearningServices/workspaces/workspace1",
			ApiVersion: "2021-07-01",
			Input: `
{
    "id": "/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/acctest5076/providers/Microsoft.MachineLearningServices/workspaces/acctest5076",
    "name": "acctest5076",
    "type": "Microsoft.MachineLearningServices/workspaces",
    "location": "westeurope",
    "tags": {},
    "etag": null,
    "properties": {
        "friendlyName": "",
        "description": "",
        "storageAccount": "/subscriptions/00000000-0000-0000-0000-00000000000/resourcegroups/acctest5076/providers/microsoft.storage/storageaccounts/acctest5076",
        "keyVault": "/subscriptions/00000000-0000-0000-0000-00000000000/resourcegroups/acctest5076/providers/microsoft.keyvault/vaults/acctest5076",
        "applicationInsights": "/subscriptions/00000000-0000-0000-0000-00000000000/resourcegroups/acctest5076/providers/microsoft.insights/components/acctest5076",
        "hbiWorkspace": false,
        "tenantId": "72f988bf-86f1-41af-91ab-2d7cd011db47",
        "imageBuildCompute": "",
        "provisioningState": "Succeeded",
        "containerRegistry": null,
        "notebookInfo": {
            "resourceId": "3fc74d5a1d3844e391497d9a71ddde42",
            "fqdn": "ml-acctest5076-westeurope-130c4f07-5479-452c-b3b8-e65da97aa0a5.westeurope.notebooks.azure.net",
            "isPrivateLinkEnabled": false,
            "notebookPreparationError": null
        },
        "storageHnsEnabled": false,
        "workspaceId": "130c4f07-5479-452c-b3b8-e65da97aa0a5",
        "linkedModelInventoryArmId": null,
        "privateLinkCount": 0,
        "allowPublicAccessWhenBehindVnet": false,
        "publicNetworkAccess": "Disabled",
        "discoveryUrl": "https://westeurope.api.azureml.ms/discovery",
        "mlFlowTrackingUri": "azureml://westeurope.api.azureml.ms/mlflow/v1.0/subscriptions/00000000-0000-0000-0000-00000000000/resourceGroups/acctest5076/providers/Microsoft.MachineLearningServices/workspaces/acctest5076",
        "sdkTelemetryAppInsightsKey": "9ac578de-874f-4fea-85bc-7e4cefd0d47f"
    },
    "identity": {
        "type": "SystemAssigned",
        "principalId": "66305a99-773c-4aa0-9b18-55b9ee5a68c0",
        "tenantId": "72f988bf-86f1-41af-91ab-2d7cd011db47"
    },
    "sku": {
        "name": "Basic",
        "tier": "Basic"
    },
    "systemData": {
        "createdAt": "2021-11-12T07:20:01.8875955Z",
        "createdBy": "someone@microsoft.com",
        "createdByType": "User",
        "lastModifiedAt": "2021-11-12T07:20:01.8875955Z",
        "lastModifiedBy": "someone@microsoft.com",
        "lastModifiedByType": "User"
    }
}
`,
			Output: `
{
    "name": "acctest5076",
    "location": "westeurope",
    "tags": {},
    "properties": {
        "friendlyName": "",
        "description": "",
        "storageAccount": "/subscriptions/00000000-0000-0000-0000-00000000000/resourcegroups/acctest5076/providers/microsoft.storage/storageaccounts/acctest5076",
        "keyVault": "/subscriptions/00000000-0000-0000-0000-00000000000/resourcegroups/acctest5076/providers/microsoft.keyvault/vaults/acctest5076",
        "applicationInsights": "/subscriptions/00000000-0000-0000-0000-00000000000/resourcegroups/acctest5076/providers/microsoft.insights/components/acctest5076",
        "hbiWorkspace": false,
        "imageBuildCompute": "",
        "containerRegistry": null,
        "allowPublicAccessWhenBehindVnet": false,
        "publicNetworkAccess": "Disabled",
        "discoveryUrl": "https://westeurope.api.azureml.ms/discovery"
    },
    "identity": {
        "type": "SystemAssigned"
    },
    "sku": {
        "name": "Basic",
        "tier": "Basic"
    }
}
`,
		},
	}

	for _, data := range testData {
		resourceType := utils.GetResourceType(data.Id)

		var input, output interface{}
		_ = json.Unmarshal([]byte(data.Input), &input)
		_ = json.Unmarshal([]byte(data.Output), &output)

		def, err := azure.GetResourceDefinition(resourceType, data.ApiVersion)
		if err != nil {
			t.Fatal(err)
		}

		if def != nil {
			res := (*def).GetWriteOnly(input)
			if !reflect.DeepEqual(res, output) {
				resJson, _ := json.Marshal(res)
				t.Errorf("expect %s got %s", data.Output, string(resJson))
			}
		} else {
			t.Fatalf("failed to load resource definition for id: %s, api-version: %s", data.Id, data.ApiVersion)
		}
	}
}
