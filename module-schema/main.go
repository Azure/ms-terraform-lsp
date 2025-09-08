package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/ms-terraform-lsp/module-schema/processor"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fetch := flag.Bool("fetch", false, "Fetch module repositories from GitHub (requires GITHUB_TOKEN)")
	dir := flag.String("dir", "fetched_hcl_files", "Directory containing combinedVar, examples, and readmes (ignored when -fetch is used)")
	noProcess := flag.Bool("no-process", false, "Skip processing after fetch (only valid with -fetch)")
	flag.Parse()

	// If fetch requested, perform fetch + combine using the internal fixed outputDir (fetched_hcl_files)
	if *fetch {
		fmt.Fprintln(os.Stderr, "[module-schema] Fetching repositories (this may take several minutes)...")
		processor.FetchRepositoryData()
		fmt.Fprintln(os.Stderr, "[module-schema] Combining variable files...")
		if err := processor.CombineVariableFiles(); err != nil {
			return fmt.Errorf("combine variable files: %w", err)
		}
		// Force dir to the processor's outputDir constant
		*dir = "fetched_hcl_files"
		if *noProcess {
			fmt.Fprintln(os.Stderr, "[module-schema] Fetch & combine complete. Skipping processing per --no-process.")
			return nil
		}
	}

	// Validate expected structure before processing
	combinedVarPath := filepath.Join(*dir, "combinedVar")
	if _, err := os.Stat(combinedVarPath); os.IsNotExist(err) {
		return fmt.Errorf("expected directory %s not found. Run with -fetch first or prepare directory structure: %s/{combinedVar,examples,readmes}", combinedVarPath, *dir)
	}

	_, err := processor.ProcessBatchOutput(*dir)
	if err != nil {
		return fmt.Errorf("error processing output: %v", err)
	}
	return nil
}
