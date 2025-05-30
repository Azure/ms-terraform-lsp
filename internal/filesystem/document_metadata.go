package filesystem

import (
	"sync"

	"github.com/Azure/ms-terraform-lsp/internal/source"
)

type documentMetadata struct {
	dh DocumentHandler

	mu      *sync.RWMutex
	isOpen  bool
	version int
	langId  string
	lines   source.Lines
}

func NewDocumentMetadata(dh DocumentHandler, langId string, content []byte) *documentMetadata {
	return &documentMetadata{
		dh:     dh,
		mu:     &sync.RWMutex{},
		langId: langId,
		lines:  source.MakeSourceLines(dh.Filename(), content),
	}
}

func (d *documentMetadata) setOpen(isOpen bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.isOpen = isOpen
}

func (d *documentMetadata) setVersion(version int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.version = version
}

func (d *documentMetadata) updateLines(content []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lines = source.MakeSourceLines(d.dh.Filename(), content)
}

func (d *documentMetadata) Lines() source.Lines {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lines
}

func (d *documentMetadata) Version() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.version
}

func (d *documentMetadata) IsOpen() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isOpen
}
