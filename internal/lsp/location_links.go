package lsp

import (
	"path/filepath"

	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	"github.com/Azure/ms-terraform-lsp/internal/uri"
	"github.com/hashicorp/hcl-lang/decoder"
)

func RefTargetsToLocationLinks(targets decoder.ReferenceTargets, linkSupport bool) interface{} {
	if linkSupport {
		links := make([]lsp.LocationLink, 0)
		for _, target := range targets {
			links = append(links, refTargetToLocationLink(target))
		}
		return links
	}

	locations := make([]lsp.Location, 0)
	for _, target := range targets {
		locations = append(locations, refTargetToLocation(target))
	}
	return locations
}

func refTargetToLocationLink(target *decoder.ReferenceTarget) lsp.LocationLink {
	targetUri := uri.FromPath(filepath.Join(target.Path.Path, target.Range.Filename))

	locLink := lsp.LocationLink{
		OriginSelectionRange: HCLRangeToLSP(target.OriginRange),
		TargetURI:            lsp.DocumentURI(targetUri),
		TargetRange:          HCLRangeToLSP(target.Range),
	}

	if target.DefRangePtr != nil {
		locLink.TargetSelectionRange = HCLRangeToLSP(*target.DefRangePtr)
	}

	return locLink
}

func refTargetToLocation(target *decoder.ReferenceTarget) lsp.Location {
	targetUri := uri.FromPath(filepath.Join(target.Path.Path, target.Range.Filename))

	return lsp.Location{
		URI:   lsp.DocumentURI(targetUri),
		Range: HCLRangeToLSP(target.Range),
	}
}
