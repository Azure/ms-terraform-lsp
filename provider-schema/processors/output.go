package processors

import (
	"encoding/json"
	"fmt"
	"os"
)

// ProcessOutput combines schema and markdown outputs and saves the result to a file
func ProcessOutput(providerPath, gitBranch, outputFile string) (TerraformObjects, error) {
	//// if outputFile already exists, try to read it and use that instead
	//if _, err := os.Stat(outputFile); err == nil {
	//	// File exists, read it
	//	data, err := os.ReadFile(outputFile)
	//	if err != nil {
	//		return nil, fmt.Errorf("error reading existing output file: %v", err)
	//	}
	//
	//	var existingOutput TerraformObjects
	//	err = json.Unmarshal(data, &existingOutput)
	//	if err != nil {
	//		return nil, fmt.Errorf("error unmarshaling existing output file: %v", err)
	//	}
	//
	//	fmt.Printf("Using existing output from %s\n", outputFile)
	//	return existingOutput, nil
	//} else if !os.IsNotExist(err) {
	//	return nil, fmt.Errorf("error checking output file: %v", err)
	//}

	// Step 1: Generate schema
	schemaOutput, err := ProcessSchema(providerPath, gitBranch)
	if err != nil {
		return nil, fmt.Errorf("error processing schema: %v", err)
	}

	// Step 2: Generate markdown docs
	markdownOutput, err := ProcessMarkdown(providerPath)
	if err != nil {
		return nil, fmt.Errorf("error processing markdown: %v", err)
	}

	// Step 3: Combine schema and markdown outputs
	combinedOutput, err := CombineSchemaAndMarkdown(schemaOutput, markdownOutput)
	if err != nil {
		return nil, fmt.Errorf("error combining outputs: %v", err)
	}

	// Step 4: Marshal combined output to JSON
	jsonOutput, err := json.MarshalIndent(combinedOutput, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling combined output: %v", err)
	}

	// Step 5: Write combined output to a file
	err = os.WriteFile(outputFile, jsonOutput, 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing combined output to file: %v", err)
	}

	fmt.Printf("Combined output successfully written to %s\n", outputFile)
	return combinedOutput, nil
}
