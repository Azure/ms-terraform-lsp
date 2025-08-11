package main

import (
	"fmt"
	"github.com/Azure/ms-terraform-lsp/module-schema/processor"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	//processor.FetchRepositoryData()
	_, err := processor.ProcessBatchOutput("/outputDir")
	if err != nil {
		return fmt.Errorf("error processing output: %v", err)
	}

	return nil
}
