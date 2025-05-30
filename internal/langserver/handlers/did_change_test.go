package handlers

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Azure/ms-terraform-lsp/internal/filesystem"
	"github.com/Azure/ms-terraform-lsp/internal/langserver"
	"github.com/Azure/ms-terraform-lsp/internal/lsp"
	"github.com/google/go-cmp/cmp"
)

func TestLangServer_didChange_sequenceOfPartialChanges(t *testing.T) {
	tmpDir := TempDir(t)

	fs := filesystem.NewFilesystem()
	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
		Filesystem: fs,
	}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
	    "capabilities": {},
	    "rootUri": %q,
	    "processId": 12345
	}`, tmpDir.URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})

	originalText := `variable "service_host" {
  default = "blah"
}

module "app" {
  source = "./sub"
  service_listeners = [
    {
      hosts    = [var.service_host]
      listener = ""
    }
  ]
}
`
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
    "textDocument": {
        "languageId": "terraform",
        "version": 0,
        "uri": "%s/main.tf",
        "text": %q
    }
}`, TempDir(t).URI(), originalText),
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didChange",
		ReqParams: fmt.Sprintf(`{
    "textDocument": {
        "version": 1,
        "uri": "%s/main.tf"
    },
    "contentChanges": [
        {
            "text": "\n",
            "rangeLength": 0,
            "range": {
                "end": {
                    "line": 8,
                    "character": 18
                },
                "start": {
                    "line": 8,
                    "character": 18
                }
            }
        },
        {
            "text": "      ",
            "rangeLength": 0,
            "range": {
                "end": {
                    "line": 9,
                    "character": 0
                },
                "start": {
                    "line": 9,
                    "character": 0
                }
            }
        }
    ]
}`, TempDir(t).URI()),
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didChange",
		ReqParams: fmt.Sprintf(`{
    "textDocument": {
        "version": 2,
        "uri": "%s/main.tf"
    },
    "contentChanges": [
        {
            "text": "  ",
            "rangeLength": 0,
            "range": {
                "end": {
                    "line": 9,
                    "character": 6
                },
                "start": {
                    "line": 9,
                    "character": 6
                }
            }
        }
    ]
}`, TempDir(t).URI()),
	})

	path := filepath.Join(TempDir(t).Dir(), "main.tf")
	doc, err := fs.GetDocument(lsp.FileHandlerFromPath(path))
	if err != nil {
		t.Fatal(err)
	}
	text, err := doc.Text()
	if err != nil {
		t.Fatal(err)
	}
	expectedText := `variable "service_host" {
  default = "blah"
}

module "app" {
  source = "./sub"
  service_listeners = [
    {
      hosts    = [
        var.service_host]
      listener = ""
    }
  ]
}
`

	if diff := cmp.Diff(expectedText, string(text)); diff != "" {
		t.Fatalf("unexpected text: %s", diff)
	}
}
