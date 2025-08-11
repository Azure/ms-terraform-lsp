package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SearchResult defines the structure for the JSON response from the registry's module search API.
type SearchResult struct {
	Modules []struct {
		ID string `json:"id"`
	} `json:"modules"`
}

// GetModuleByRepoName queries the HashiCorp Registry to find a module name based on a GitHub repository name.
func GetModuleByRepoName(repoName string) (string, error) {
	// The search endpoint for the registry.
	const baseURL = "https://registry.terraform.io/v1/modules/search"
	const repoPrefix = "terraform-azurerm-"

	registryName, hasPrefix := strings.CutPrefix(repoName, repoPrefix)
	if !hasPrefix {
		return "", fmt.Errorf("repository name must start with '%s'", repoPrefix)
	}
	// URL-encode the repository name to ensure it's a valid query parameter.
	query := url.QueryEscape(registryName)
	requestURL := fmt.Sprintf("%s?q=%s", baseURL, query)

	// Perform the GET request.
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("request to registry failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry returned a non-200 status: %d", resp.StatusCode)
	}

	// Decode the JSON response into our struct.
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// If no modules were found, return an error.
	if len(result.Modules) == 0 {
		return "", fmt.Errorf("no modules found for repo: %s", repoName)
	}

	// The most relevant result is typically the first one.
	return result.Modules[0].ID, nil
}
