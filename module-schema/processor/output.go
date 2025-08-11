package processor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/Azure/ms-terraform-lsp/module-schema/schema"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const outputFileName = "output.json"

// CombineVariableFiles combines all {repo_name}_variables.*.tf files into {repo_name}_variables.tf under fetched_hcl_files/combinedVar
func CombineVariableFiles() error {
	variablesDir := filepath.Join(outputDir, "variables")
	combinedVarDir := filepath.Join(outputDir, "combinedVar")

	if err := os.MkdirAll(combinedVarDir, 0755); err != nil {
		return fmt.Errorf("error creating combinedVar output directory: %w", err)
	}

	files, err := os.ReadDir(variablesDir)
	if err != nil {
		return fmt.Errorf("error reading variables directory: %w", err)
	}

	repoFiles := make(map[string][]string)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tf") {
			continue
		}
		fileName := file.Name()
		if !strings.Contains(fileName, "_variables.") {
			continue
		}
		parts := strings.SplitN(fileName, "_variables.", 2)
		if len(parts) != 2 {
			log.Printf("Skipping file with unexpected naming pattern: %s", fileName)
			continue
		}
		repoName := parts[0]
		repoFiles[repoName] = append(repoFiles[repoName], fileName)
	}

	for repoName, fileNames := range repoFiles {
		sort.Strings(fileNames)
		var combinedContent strings.Builder
		combinedContent.WriteString(fmt.Sprintf("# Combined variables file for repository: %s\n", repoName))
		combinedContent.WriteString(fmt.Sprintf("# Generated on: %s\n", time.Now().Format("2006-01-02 15:04:05")))
		combinedContent.WriteString(fmt.Sprintf("# Source files: %s\n\n", strings.Join(fileNames, ", ")))
		for i, fileName := range fileNames {
			sourcePath := filepath.Join(variablesDir, fileName)
			content, err := os.ReadFile(sourcePath)
			if err != nil {
				log.Printf("Error reading file '%s': %v", fileName, err)
				continue
			}
			if i > 0 {
				combinedContent.WriteString("\n# " + strings.Repeat("=", 80) + "\n")
			}
			combinedContent.WriteString(fmt.Sprintf("# Source: %s\n", fileName))
			combinedContent.WriteString("# " + strings.Repeat("=", 80) + "\n\n")
			combinedContent.Write(content)
			if !strings.HasSuffix(string(content), "\n") {
				combinedContent.WriteString("\n")
			}
			combinedContent.WriteString("\n")
		}
		targetPath := filepath.Join(combinedVarDir, fmt.Sprintf("%s_variables.tf", repoName))
		err := os.WriteFile(targetPath, []byte(combinedContent.String()), 0644)
		if err != nil {
			log.Printf("Error writing combined file '%s': %v", targetPath, err)
			continue
		}
		log.Printf("Successfully combined %d files into: %s", len(fileNames), targetPath)
	}
	return nil
}

func ProcessBatchOutput(dirPath string) ([]schema.TerraformObject, error) {
	terraformObjects, err := processVariableFiles(dirPath)
	if err != nil {
		return nil, err
	}

	if err := writeJSONOutput(terraformObjects); err != nil {
		return nil, err
	}

	fmt.Printf("Output written to %s\n", outputFileName)
	return terraformObjects, nil
}

// processVariableFiles processes all variable files and creates TerraformObject instances
func processVariableFiles(dirPath string) ([]schema.TerraformObject, error) {
	var terraformObjects []schema.TerraformObject
	variablesDir := filepath.Join(dirPath, "combinedVar")
	examplesDir := filepath.Join(dirPath, "examples")
	readmesDir := filepath.Join(dirPath, "readmes")

	err := filepath.WalkDir(variablesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through directory: %w", err)
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), "_variables.tf") {
			return nil
		}

		terraformObj, err := processSingleVariableFile(path, examplesDir, readmesDir)
		if err != nil {
			return err
		}

		terraformObjects = append(terraformObjects, terraformObj)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return terraformObjects, nil
}

// processSingleVariableFile processes a single variable file and returns a TerraformObject
func processSingleVariableFile(variableFilePath, examplesDir, readmesDir string) (schema.TerraformObject, error) {
	objName := strings.TrimSuffix(filepath.Base(variableFilePath), "_variables.tf")

	// Parse the variable file
	parsedBlock, err := parseVariableFile(variableFilePath, objName)
	if err != nil {
		return schema.TerraformObject{}, err
	}

	// Read example HCL content
	exampleHcl, err := readExampleContent(examplesDir, objName)
	if err != nil {
		return schema.TerraformObject{}, err
	}

	// Read README content
	details, err := readReadmeContent(readmesDir, objName)
	if err != nil {
		return schema.TerraformObject{}, err
	}

	return schema.TerraformObject{
		Name:       "Azure/" + objName + "/azurerm",
		Fields:     parsedBlock.Attributes,
		ExampleHCL: exampleHcl,
		Details:    details,
	}, nil
}

// parseVariableFile parses a variable file and returns the parsed block
func parseVariableFile(variableFilePath, objName string) (*schema.SchemaBlock, error) {
	hclBlock, err := schema.FromHCLFile(variableFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL file %s: %w", variableFilePath, err)
	}

	parsedBlock, err := FindFieldsFromBlock(hclBlock, "Azure/"+objName+"/azurerm")
	if err != nil {
		return nil, fmt.Errorf("failed to find fields in block for %s: %w", variableFilePath, err)
	}

	return parsedBlock, nil
}

// readExampleContent reads and processes example HCL content
func readExampleContent(examplesDir, objName string) (string, error) {
	exampleFilePath := filepath.Join(examplesDir, fmt.Sprintf("%s_example.tf", objName))
	exampleHclContent, err := os.ReadFile(exampleFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Return empty string if file doesn't exist
		}
		return "", fmt.Errorf("error reading example file %s: %w", exampleFilePath, err)
	}

	exampleHcl := string(exampleHclContent)
	// Replace 'source = "../../"' with the correct value if present (handle variable spacing)
	sourceRegex := regexp.MustCompile(`source\s*=\s*"\.\.\/\.\.\/?"`)
	exampleHcl = sourceRegex.ReplaceAllString(exampleHcl, `source = "Azure/`+objName+`/azurerm"`)

	return exampleHcl, nil
}

// readReadmeContent reads README.md content for details
func readReadmeContent(readmesDir, objName string) (string, error) {
	readmeFilePath := filepath.Join(readmesDir, fmt.Sprintf("%s_README.md", objName))
	readmeContent, err := os.ReadFile(readmeFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Return empty string if file doesn't exist
		}
		return "", fmt.Errorf("error reading README file %s: %w", readmeFilePath, err)
	}

	return string(readmeContent), nil
}

// writeJSONOutput marshals and writes the terraform objects to JSON file
func writeJSONOutput(terraformObjects []schema.TerraformObject) error {
	jsonOutput, err := json.MarshalIndent(terraformObjects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal terraform objects to JSON: %w", err)
	}

	err = os.WriteFile(outputFileName, jsonOutput, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output to file: %w", err)
	}

	return nil
}
