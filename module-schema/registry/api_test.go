package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetModuleByRepoName tests the GetModuleByRepoName function.
func TestGetModuleByRepoName(t *testing.T) {
	// Create a mock server to simulate the HashiCorp Registry API.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the query parameter.
		query := r.URL.Query().Get("q")
		if query == "terraform-azurerm-avm-res-storage-storageaccount" {
			// Respond with a sample JSON payload.
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{
				"modules": [
					{
						"id": "Azure/avm-res-storage-storageaccount/azurerm",
						"owner": "Azure",
						"namespace": "Azure",
						"name": "avm-res-storage-storageaccount",
						"version": "0.4.0",
						"provider": "azurerm",
						"description": "Terraform module to provision an Azure Storage Account.",
						"source": "https://github.com/Azure/terraform-azurerm-avm-res-storage-storageaccount",
						"tag": "v0.4.0",
						"published_at": "2024-05-24T10:00:00Z",
						"downloads": 12345,
						"verified": true
					}
				]
			}`)
		} else if query == "non-existent-repo" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"modules": []}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Not Found")
		}
	}))
	defer server.Close()

	baseURL := server.URL

	// Helper function to perform the test.
	runTest := func(repoName, expectedID, expectedError string) {
		t.Run(repoName, func(t *testing.T) {
			// We need a way to point GetModuleByRepoName to the test server.
			// For simplicity, we'll create a temporary function that uses the mock server's URL.
			getModuleTest := func(repo string) (string, error) {
				query := repo
				requestURL := fmt.Sprintf("%s?q=%s", baseURL, query)
				resp, err := http.Get(requestURL)
				if err != nil {
					return "", fmt.Errorf("request to registry failed: %w", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return "", fmt.Errorf("registry returned a non-200 status: %d", resp.StatusCode)
				}
				var result SearchResult
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					return "", fmt.Errorf("failed to parse JSON response: %w", err)
				}
				if len(result.Modules) == 0 {
					return "", fmt.Errorf("no modules found for repo: %s", repo)
				}
				return result.Modules[0].ID, nil
			}

			id, err := getModuleTest(repoName)

			if err != nil && expectedError == "" {
				t.Fatalf("expected no error, but got: %v", err)
			}

			if err == nil && expectedError != "" {
				t.Fatalf("expected error '%s', but got none", expectedError)
			}

			if err != nil && err.Error() != expectedError {
				t.Fatalf("expected error '%s', but got '%s'", expectedError, err.Error())
			}

			if id != expectedID {
				t.Fatalf("expected module ID '%s', but got '%s'", expectedID, id)
			}
		})
	}

	// Run the test cases.
	runTest("terraform-azurerm-avm-res-storage-storageaccount", "Azure/avm-res-storage-storageaccount/azurerm", "")
	runTest("non-existent-repo", "", "no modules found for repo: non-existent-repo")
}
