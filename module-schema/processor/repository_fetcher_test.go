package processor

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v63/github"
)

type mockFile struct {
	name    string
	content string
}

// createTestGitHubClient creates a test GitHub client with a mock server
func createTestGitHubClient(handler http.HandlerFunc) (*github.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := github.NewClient(&http.Client{})

	// Parse the server URL and set it as the base URL with trailing slash
	serverURL, _ := url.Parse(server.URL + "/")
	client.BaseURL = serverURL

	return client, server
}

func TestFetchAndSaveFile(t *testing.T) {
	testCases := []struct {
		name          string
		filePath      string
		fileContent   string
		expectedError bool
		errorMessage  string
		statusCode    int
	}{
		{
			name:          "successful_file_fetch",
			filePath:      "test.tf",
			fileContent:   "variable \"test\" {\n  type = string\n}",
			expectedError: false,
		},
		{
			name:          "file_not_found",
			filePath:      "nonexistent.tf",
			expectedError: true,
			errorMessage:  "file 'nonexistent.tf' not found",
			statusCode:    404,
		},
		{
			name:          "empty_content",
			filePath:      "empty.tf",
			fileContent:   "",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary output directory
			tempDir := t.TempDir()
			outputFilePath := filepath.Join(tempDir, "output.tf")

			// Create mock server
			handler := func(w http.ResponseWriter, r *http.Request) {
				if tc.statusCode == 404 {
					w.WriteHeader(404)
					fmt.Fprint(w, `{"message": "Not Found"}`)
					return
				}

				// Encode content as base64 (like GitHub API)
				encodedContent := base64.StdEncoding.EncodeToString([]byte(tc.fileContent))

				response := fmt.Sprintf(`{
					"name": "%s",
					"path": "%s",
					"content": "%s",
					"encoding": "base64"
				}`, filepath.Base(tc.filePath), tc.filePath, encodedContent)

				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, response)
			}

			client, server := createTestGitHubClient(handler)
			defer server.Close()

			// Create a mock repository
			repo := &github.Repository{
				Name:          github.String("test-repo"),
				DefaultBranch: github.String("main"),
			}

			// Test the function
			err := fetchAndSaveFile(context.Background(), client, "testorg", repo, tc.filePath, outputFilePath)

			// Check error expectations
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMessage != "" && !strings.Contains(err.Error(), tc.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tc.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Verify file was created and content is correct
				if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
					t.Errorf("Expected output file to be created")
				} else {
					content, err := os.ReadFile(outputFilePath)
					if err != nil {
						t.Fatalf("Failed to read output file: %v", err)
					}
					if string(content) != tc.fileContent {
						t.Errorf("Expected content '%s', got '%s'", tc.fileContent, string(content))
					}
				}
			}
		})
	}
}

func TestFetchAndSaveAllVariableFiles(t *testing.T) {
	testCases := []struct {
		name          string
		repoFiles     []mockFile
		expectedFiles int
		expectedError bool
		errorMessage  string
	}{
		{
			name: "multiple_variable_files",
			repoFiles: []mockFile{
				{name: "variables.tf", content: "variable \"test1\" { type = string }"},
				{name: "variables.auto.tf", content: "variable \"test2\" { type = string }"},
				{name: "main.tf", content: "# main file"},
				{name: "README.md", content: "# Test repo"},
			},
			expectedFiles: 2,
			expectedError: false,
		},
		{
			name: "single_variable_file",
			repoFiles: []mockFile{
				{name: "variables.tf", content: "variable \"test\" { type = string }"},
				{name: "main.tf", content: "# main file"},
			},
			expectedFiles: 1,
			expectedError: false,
		},
		{
			name: "no_variable_files",
			repoFiles: []mockFile{
				{name: "main.tf", content: "# main file"},
				{name: "README.md", content: "# Test repo"},
			},
			expectedFiles: 0,
			expectedError: true,
			errorMessage:  "no variable files found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary output directory
			tempDir := t.TempDir()

			// Create mock server that handles both directory listing and individual file requests
			handler := func(w http.ResponseWriter, r *http.Request) {
				// Check if this is a request for the root directory (ends with /contents/)
				if strings.HasSuffix(r.URL.Path, "/contents/") || strings.HasSuffix(r.URL.Path, "/contents") {
					// Build JSON response for directory listing
					var fileItems []string
					for _, file := range tc.repoFiles {
						fileItem := fmt.Sprintf(`{
							"name": "%s",
							"path": "%s",
							"type": "file"
						}`, file.name, file.name)
						fileItems = append(fileItems, fileItem)
					}

					response := "[" + strings.Join(fileItems, ",") + "]"
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, response)
					return
				}

				// Handle individual file requests
				fileName := filepath.Base(r.URL.Path)
				for _, file := range tc.repoFiles {
					if file.name == fileName {
						encodedContent := base64.StdEncoding.EncodeToString([]byte(file.content))
						response := fmt.Sprintf(`{
							"name": "%s",
							"path": "%s",
							"content": "%s",
							"encoding": "base64"
						}`, file.name, file.name, encodedContent)

						w.Header().Set("Content-Type", "application/json")
						fmt.Fprint(w, response)
						return
					}
				}

				// File not found
				w.WriteHeader(404)
				fmt.Fprint(w, `{"message": "Not Found"}`)
			}

			client, server := createTestGitHubClient(handler)
			defer server.Close()

			// Create a mock repository
			repo := &github.Repository{
				Name:          github.String("test-repo"),
				DefaultBranch: github.String("main"),
			}

			// Test the function
			savedFiles, err := fetchAndSaveAllVariableFiles(context.Background(), client, "testorg", repo, tempDir, "test-module")

			// Check error expectations
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMessage != "" && !strings.Contains(err.Error(), tc.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tc.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Check number of saved files
				if len(savedFiles) != tc.expectedFiles {
					t.Errorf("Expected %d saved files, got %d", tc.expectedFiles, len(savedFiles))
				}

				// Verify files were actually created
				for _, savedFile := range savedFiles {
					if _, err := os.Stat(savedFile); os.IsNotExist(err) {
						t.Errorf("Expected file to be created: %s", savedFile)
					}
				}
			}
		})
	}
}

func TestFetchAndSaveFile_InvalidBase64(t *testing.T) {
	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "output.tf")

	// Create mock server that returns invalid base64
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"name": "test.tf",
			"path": "test.tf",
			"content": "invalid-base64-content!@#$%",
			"encoding": "base64"
		}`

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, response)
	}

	client, server := createTestGitHubClient(handler)
	defer server.Close()

	repo := &github.Repository{
		Name:          github.String("test-repo"),
		DefaultBranch: github.String("main"),
	}

	err := fetchAndSaveFile(context.Background(), client, "testorg", repo, "test.tf", outputFilePath)

	if err == nil {
		t.Error("Expected error due to invalid base64 content")
	}
	if !strings.Contains(err.Error(), "error decoding content") {
		t.Errorf("Expected error about decoding content, got: %v", err)
	}
}

func TestFetchAndSaveFile_NoContent(t *testing.T) {
	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "output.tf")

	// Create mock server that returns null content (directory)
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"name": "test-dir",
			"path": "test-dir",
			"type": "dir"
		}`

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, response)
	}

	client, server := createTestGitHubClient(handler)
	defer server.Close()

	repo := &github.Repository{
		Name:          github.String("test-repo"),
		DefaultBranch: github.String("main"),
	}

	err := fetchAndSaveFile(context.Background(), client, "testorg", repo, "test-dir", outputFilePath)

	if err == nil {
		t.Error("Expected error due to no content (directory)")
	}
	if !strings.Contains(err.Error(), "has no content or is a directory") {
		t.Errorf("Expected error about no content, got: %v", err)
	}
}

func TestFetchAndSaveFile_WriteError(t *testing.T) {
	// Try to write to an invalid path (should cause write error)
	invalidPath := "/root/readonly/output.tf" // This should fail on most systems

	handler := func(w http.ResponseWriter, r *http.Request) {
		encodedContent := base64.StdEncoding.EncodeToString([]byte("test content"))
		response := fmt.Sprintf(`{
			"name": "test.tf",
			"path": "test.tf",
			"content": "%s",
			"encoding": "base64"
		}`, encodedContent)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, response)
	}

	client, server := createTestGitHubClient(handler)
	defer server.Close()

	repo := &github.Repository{
		Name:          github.String("test-repo"),
		DefaultBranch: github.String("main"),
	}

	err := fetchAndSaveFile(context.Background(), client, "testorg", repo, "test.tf", invalidPath)

	if err == nil {
		t.Error("Expected error due to write failure")
	}
	if !strings.Contains(err.Error(), "error writing file") {
		t.Errorf("Expected error about writing file, got: %v", err)
	}
}

func TestFetchAndSaveAllVariableFiles_APIError(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock server that returns an API error
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		fmt.Fprint(w, `{"message": "Internal Server Error"}`)
	}

	client, server := createTestGitHubClient(handler)
	defer server.Close()

	repo := &github.Repository{
		Name:          github.String("test-repo"),
		DefaultBranch: github.String("main"),
	}

	savedFiles, err := fetchAndSaveAllVariableFiles(context.Background(), client, "testorg", repo, tempDir, "test-module")

	if err == nil {
		t.Error("Expected error due to API error")
	}
	if !strings.Contains(err.Error(), "error fetching root directory content") {
		t.Errorf("Expected error about fetching directory content, got: %v", err)
	}
	if len(savedFiles) != 0 {
		t.Errorf("Expected no saved files on error, got %d", len(savedFiles))
	}
}

func TestFetchAndSaveFile_NetworkError(t *testing.T) {
	// Create a client that points to a non-existent server
	client := github.NewClient(&http.Client{})
	invalidURL, _ := url.Parse("http://localhost:99999") // Non-existent port
	client.BaseURL = invalidURL

	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "output.tf")

	repo := &github.Repository{
		Name:          github.String("test-repo"),
		DefaultBranch: github.String("main"),
	}

	err := fetchAndSaveFile(context.Background(), client, "testorg", repo, "test.tf", outputFilePath)

	if err == nil {
		t.Error("Expected error due to network failure")
	}
	// The exact error message may vary, but it should be about fetching the file
	if !strings.Contains(err.Error(), "error fetching file") {
		t.Errorf("Expected error about fetching file, got: %v", err)
	}
}

// Benchmark tests for performance
func BenchmarkFetchAndSaveFile(b *testing.B) {
	tempDir := b.TempDir()

	// Create a simple mock server
	handler := func(w http.ResponseWriter, r *http.Request) {
		content := "variable \"test\" { type = string }"
		encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
		response := fmt.Sprintf(`{
			"name": "test.tf",
			"path": "test.tf",
			"content": "%s",
			"encoding": "base64"
		}`, encodedContent)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, response)
	}

	client, server := createTestGitHubClient(handler)
	defer server.Close()

	repo := &github.Repository{
		Name:          github.String("test-repo"),
		DefaultBranch: github.String("main"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tempDir, fmt.Sprintf("output_%d.tf", i))
		err := fetchAndSaveFile(context.Background(), client, "testorg", repo, "test.tf", outputPath)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
