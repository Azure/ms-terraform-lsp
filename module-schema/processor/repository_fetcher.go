package processor

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Azure/ms-terraform-lsp/module-schema/registry"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

const (
	githubOrg       = "Azure"
	repoNamePattern = "^terraform-azurerm-avm-.*$" // Adjust this regex to match your repository naming convention
	outputDir       = "fetched_hcl_files"          // Directory to save fetched files
)

func FetchRepositoryData() {
	client, err := setupGitHubClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := createOutputDirectories(); err != nil {
		log.Fatal(err)
	}

	repos, err := fetchAllRepositories(client)
	if err != nil {
		log.Fatal(err)
	}

	failedExampleRepos := processRepositories(client, repos)

	if err := writeFailedExampleReposReport(failedExampleRepos); err != nil {
		log.Printf("Error writing failed repos report: %v", err)
	}
}

// setupGitHubClient initializes and returns a GitHub client with authentication
func setupGitHubClient() (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set. Please set it to your GitHub Personal Access Token")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
}

// createOutputDirectories creates all necessary output directories
func createOutputDirectories() error {
	variablesOutputDir := filepath.Join(outputDir, "variables")
	examplesOutputDir := filepath.Join(outputDir, "examples")
	readmesOutputDir := filepath.Join(outputDir, "readmes")

	directories := []string{variablesOutputDir, examplesOutputDir, readmesOutputDir}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating output directory %s: %w", dir, err)
		}
	}

	return nil
}

// fetchAllRepositories retrieves all repositories matching the pattern from GitHub
func fetchAllRepositories(client *github.Client) ([]*github.Repository, error) {
	repoRegex, err := regexp.Compile(repoNamePattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex for repo name pattern: %w", err)
	}

	ctx := context.Background()
	log.Printf("Searching for repositories matching pattern '%s' in organization/user '%s'...", repoNamePattern, githubOrg)

	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.List(ctx, githubOrg, opt)
		if err != nil {
			return nil, fmt.Errorf("error listing repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.Printf("Found %d repositories. Filtering by pattern...", len(allRepos))

	// Filter repositories by pattern
	var matchingRepos []*github.Repository
	for _, repo := range allRepos {
		if repoRegex.MatchString(repo.GetName()) {
			matchingRepos = append(matchingRepos, repo)
		}
	}

	return matchingRepos, nil
}

// processRepositories processes each repository to fetch variables, examples, and README files
func processRepositories(client *github.Client, repos []*github.Repository) []string {
	var failedExampleRepos []string
	ctx := context.Background()

	variablesOutputDir := filepath.Join(outputDir, "variables")
	examplesOutputDir := filepath.Join(outputDir, "examples")
	readmesOutputDir := filepath.Join(outputDir, "readmes")

	for _, repo := range repos {
		log.Printf("Repository '%s' matches pattern. Processing...", repo.GetName())

		registryName, err := getRegistryName(repo.GetName())
		if err != nil {
			log.Printf("Error retrieving module name for repo '%s': %v", repo.GetName(), err)
			continue
		}

		// Fetch variables files
		if err := fetchVariableFiles(ctx, client, repo, variablesOutputDir, registryName); err != nil {
			log.Printf("Could not fetch any variable files for repo '%s', skipping: %v", repo.GetName(), err)
			continue
		}

		// Fetch example files
		if failureReason := fetchExampleFiles(ctx, client, repo, examplesOutputDir, registryName); failureReason != "" {
			failedExampleRepos = append(failedExampleRepos, fmt.Sprintf("%s - %s", repo.GetName(), failureReason))
		}

		// Fetch README files
		fetchReadmeFile(ctx, client, repo, readmesOutputDir, registryName)
	}

	return failedExampleRepos
}

// getRegistryName extracts the registry name from the repository name
func getRegistryName(repoName string) (string, error) {
	moduleName, err := registry.GetModuleByRepoName(repoName)
	if err != nil {
		return "", err
	}

	nameArr := strings.Split(moduleName, "/")
	fmt.Printf("Module name for repo '%s': %s\n", repoName, moduleName)

	if len(nameArr) < 2 {
		return "", fmt.Errorf("invalid module name format: %s", moduleName)
	}

	return nameArr[1], nil
}

// fetchVariableFiles fetches all variable files for a repository
func fetchVariableFiles(ctx context.Context, client *github.Client, repo *github.Repository, outputDir, registryName string) error {
	_, err := fetchAndSaveAllVariableFiles(ctx, client, githubOrg, repo, outputDir, registryName)
	return err
}

// fetchExampleFiles fetches example files for a repository, returns failure reason if unsuccessful
func fetchExampleFiles(ctx context.Context, client *github.Client, repo *github.Repository, outputDir, registryName string) string {
	exampleFilePath := "examples/default/main.tf"
	exampleOutputFilePath := filepath.Join(outputDir, fmt.Sprintf("%s_example.tf", registryName))

	err := fetchAndSaveFile(ctx, client, githubOrg, repo, exampleFilePath, exampleOutputFilePath)
	if err != nil {
		log.Printf("Could not fetch 'examples/default/main.tf' for repo '%s': %v", repo.GetName(), err)

		// Try fallback: find any examples/.*/main.tf
		if found, fallbackErrors := tryFallbackExamples(ctx, client, repo, exampleOutputFilePath); !found {
			log.Printf("No example main.tf found in any examples/*/ for repo '%s'", repo.GetName())

			if len(fallbackErrors) > 0 {
				return fmt.Sprintf("examples/default/main.tf not found (%v) and fallback attempts failed: %s", err, strings.Join(fallbackErrors, "; "))
			}
			return fmt.Sprintf("examples/default/main.tf not found (%v) and no subdirectories found in examples/", err)
		}
	}

	return "" // Success
}

// tryFallbackExamples attempts to find alternative example files when default is not found
func tryFallbackExamples(ctx context.Context, client *github.Client, repo *github.Repository, exampleOutputFilePath string) (bool, []string) {
	var fallbackErrors []string

	_, examplesDirContent, _, derr := client.Repositories.GetContents(
		ctx, githubOrg, repo.GetName(), "examples",
		&github.RepositoryContentGetOptions{Ref: repo.GetDefaultBranch()},
	)

	if derr != nil {
		fallbackErrors = append(fallbackErrors, fmt.Sprintf("examples directory inaccessible: %v", derr))
		return false, fallbackErrors
	}

	for _, entry := range examplesDirContent {
		if entry.GetType() == "dir" {
			subMainPath := path.Join("examples", entry.GetName(), "main.tf")
			subErr := fetchAndSaveFile(ctx, client, githubOrg, repo, subMainPath, exampleOutputFilePath)
			if subErr == nil {
				log.Printf("Fetched '%s' for repo '%s' as fallback example.", subMainPath, repo.GetName())
				return true, nil
			}
			fallbackErrors = append(fallbackErrors, fmt.Sprintf("%s: %v", subMainPath, subErr))
		}
	}

	return false, fallbackErrors
}

// fetchReadmeFile fetches the README.md file for a repository
func fetchReadmeFile(ctx context.Context, client *github.Client, repo *github.Repository, outputDir, registryName string) {
	readmeFilePath := "README.md"
	readmeOutputFilePath := filepath.Join(outputDir, fmt.Sprintf("%s_README.md", registryName))

	err := fetchAndSaveFile(ctx, client, githubOrg, repo, readmeFilePath, readmeOutputFilePath)
	if err != nil {
		log.Printf("Could not fetch 'README.md' for repo '%s': %v", repo.GetName(), err)
	} else {
		log.Printf("Successfully fetched README.md for repo '%s'", repo.GetName())
	}
}

// writeFailedExampleReposReport writes a report of repositories that failed to fetch example files
func writeFailedExampleReposReport(failedExampleRepos []string) error {
	if len(failedExampleRepos) > 0 {
		failedReposFile := filepath.Join(outputDir, "failed_example_repos.txt")
		failedContent := strings.Join(failedExampleRepos, "\n")
		err := os.WriteFile(failedReposFile, []byte(failedContent), 0644)
		if err != nil {
			return fmt.Errorf("error writing failed repos file: %w", err)
		}
		log.Printf("Saved %d failed example repos to '%s'", len(failedExampleRepos), failedReposFile)
	} else {
		log.Printf("All repositories had example files successfully fetched!")
	}
	return nil
}

func fetchAndSaveAllVariableFiles(ctx context.Context, client *github.Client, org string, repo *github.Repository, outputDir string, registryName string) ([]string, error) {
	log.Printf("Fetching directory content for repo: %s", repo.GetName())
	_, directoryContent, _, err := client.Repositories.GetContents(
		ctx,
		org,
		repo.GetName(),
		"", // Get root directory content
		&github.RepositoryContentGetOptions{Ref: repo.GetDefaultBranch()},
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching root directory content: %w", err)
	}

	var savedFiles []string
	log.Printf("Found %d files in root directory of %s", len(directoryContent), repo.GetName())
	for _, file := range directoryContent {
		log.Printf("Checking file: %s", file.GetName())
		if strings.HasPrefix(file.GetName(), "variables.") && strings.HasSuffix(file.GetName(), ".tf") {
			log.Printf("Found matching file: %s", file.GetName())
			outputFilePath := filepath.Join(outputDir, fmt.Sprintf("%s_%s", registryName, file.GetName()))
			err := fetchAndSaveFile(ctx, client, org, repo, file.GetPath(), outputFilePath)
			if err != nil {
				log.Printf("Could not fetch file '%s': %v", file.GetName(), err)
				continue
			}
			savedFiles = append(savedFiles, outputFilePath)
		}
	}

	if len(savedFiles) == 0 {
		return nil, fmt.Errorf("no variable files found in repo '%s'", repo.GetName())
	}

	return savedFiles, nil
}

func fetchAndSaveFile(ctx context.Context, client *github.Client, org string, repo *github.Repository, filePath string, outputFilePath string) error {
	fileContent, _, resp, err := client.Repositories.GetContents(
		ctx,
		org,
		repo.GetName(),
		filePath,
		&github.RepositoryContentGetOptions{Ref: repo.GetDefaultBranch()},
	)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return fmt.Errorf("file '%s' not found", filePath)
		}
		return fmt.Errorf("error fetching file '%s': %w", filePath, err)
	}

	if fileContent.Content == nil {
		return fmt.Errorf("file '%s' has no content or is a directory", filePath)
	}

	decodedContent, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(*fileContent.Content, "\n", ""))
	if err != nil {
		return fmt.Errorf("error decoding content for '%s': %w", filePath, err)
	}

	err = os.WriteFile(outputFilePath, decodedContent, 0644)
	if err != nil {
		return fmt.Errorf("error writing file '%s': %w", outputFilePath, err)
	}

	log.Printf("Successfully fetched and saved '%s' from '%s' to '%s'", filePath, repo.GetName(), outputFilePath)
	return nil
}
