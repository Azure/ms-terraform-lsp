package validate

import (
	"github.com/Azure/ms-terraform-lsp/internal/langserver/diagnostics"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func NewDiagnostics(src []byte, filename string) diagnostics.Diagnostics {
	diags := diagnostics.NewDiagnostics()
	_, schemaDiags := ValidateFile(src, filename)
	diags.EmptyRootDiagnostic()
	validateDiags := make(map[string]hcl.Diagnostics)
	validateDiags[filename] = schemaDiags
	diags.Append("schema validate", validateDiags)
	return diags
}

func ValidateFile(src []byte, filename string) (*hcl.File, hcl.Diagnostics) {
	file, _ := hclsyntax.ParseConfig(src, filename, hcl.InitialPos)
	if file == nil {
		return nil, nil
	}
	_, isHcl := file.Body.(*hclsyntax.Body)
	if !isHcl {
		return nil, nil
	}

	return file, nil
}
