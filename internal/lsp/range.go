package lsp

import (
	"github.com/Azure/ms-terraform-lsp/internal/filesystem"
	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/hashicorp/hcl/v2"
)

func lspRangeToFsRange(rng *lsp.Range) *filesystem.Range {
	if rng == nil {
		return nil
	}

	return &filesystem.Range{
		Start: filesystem.Pos{
			Line:   int(rng.Start.Line),
			Column: int(rng.Start.Character),
		},
		End: filesystem.Pos{
			Line:   int(rng.End.Line),
			Column: int(rng.End.Character),
		},
	}
}

func HCLRangeToLSP(rng hcl.Range) lsp.Range {
	return lsp.Range{
		Start: HCLPosToLSP(rng.Start),
		End:   HCLPosToLSP(rng.End),
	}
}

func HCLPosToLSP(pos hcl.Pos) lsp.Position {
	return lsp.Position{
		// #nosec G115
		Line: uint32(pos.Line - 1),
		// #nosec G115
		Character: uint32(pos.Column - 1),
	}
}

func LSPPosToHCL(pos lsp.Position) hcl.Pos {
	return hcl.Pos{
		Line:   int(pos.Line + 1),
		Column: int(pos.Character + 1),
	}
}
