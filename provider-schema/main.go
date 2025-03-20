package provider_schema

import (
	"fmt"
	"github.com/Azure/azurerm-lsp/provider-schema/processors"
	"os"
	"os/exec"
)

func main() {
	providerPath := "/Users/harryqu/Projects-m/terraform-m"
	gitBranch := "main"
	dirPath := os.Getenv("PWD") + "/provider-schema"

	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s/processors/.tools/generate-provider-schema/run.sh %s %s", dirPath, providerPath, gitBranch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running script: %v\n", err)
		return
	}

	_, err = processors.ProcessOutput(providerPath, gitBranch, "combined_output.json")
	if err != nil {
		fmt.Printf("Error processing output: %v\n", err)
		return
	}
}
