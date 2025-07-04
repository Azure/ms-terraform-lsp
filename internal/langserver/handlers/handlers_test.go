package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Azure/ms-terraform-lsp/internal/langserver"
	"github.com/Azure/ms-terraform-lsp/internal/lsp"
)

func initializeResponse(t *testing.T, commandPrefix string) string {
	return `{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
		  "capabilities": {
			"textDocumentSync": {
			  "openClose": true,
			  "change": 2,
			  "save": {}
			},
			"completionProvider": {
			  "triggerCharacters": [
				"",
				" ",
				".",
				"/",
				"@",
				"{",
				"\""
			  ],
			  "completionItem": {}
			},
			"hoverProvider": true,
			"declarationProvider": false,
			"codeActionProvider": {
			  "codeActionKinds": [
				"refactor.rewrite"
			  ]
			},
			"executeCommandProvider": {
				"commands": [
					"ms-terraform.telemetry",
					"ms-terraform.aztfauthorize",
					"ms-terraform.convertJsonToAzapi",
					"ms-terraform.aztfmigrate"
				],
				"workDoneProgress": true
			}
		  },
		  "serverInfo": {
			"name": "azurerm-lsp"
		  }
		}
	}`
}

func TestInitalizeAndShutdown(t *testing.T) {
	tmpDir := TempDir(t)

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI()),
	}, initializeResponse(t, ""))
	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "shutdown", ReqParams: `{}`,
	},
		`{
		"jsonrpc": "2.0",
		"id": 2,
		"result": null
	}`)
}

func TestInitalizeWithCommandPrefix(t *testing.T) {
	tmpDir := TempDir(t)

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345,
		"initializationOptions": {
			"commandPrefix": "1"
		}
	}`, tmpDir.URI()),
	}, initializeResponse(t, "1"))
}

func TestEOF(t *testing.T) {
	tmpDir := TempDir(t)

	ms := newMockSession(&MockSessionInput{})
	ls := langserver.NewLangServerMock(t, ms.new)
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI()),
	}, initializeResponse(t, ""))

	ls.CloseClientStdout(t)

	// Session is stopped after all other operations stop
	// which may take some time
	time.Sleep(250 * time.Millisecond)

	if !ms.StopFuncCalled() {
		t.Fatal("Expected session to stop on EOF")
	}
	if ls.StopFuncCalled() {
		t.Fatal("Expected server not to stop on EOF")
	}
}

// TempDir creates a temporary directory containing the test name, as well any
// additional nested dir specified, use slash "/" to nest for more complex
// setups
//
//	ex: TempDir(t, "a/b", "c")
//	├── a
//	│   └── b
//	└── c
//
// The returned filehandler is the parent tmp dir
func TempDir(t *testing.T, nested ...string) lsp.FileHandler {
	tmpDir := filepath.Join(os.TempDir(), "azurerm-lsp", t.Name())
	err := os.MkdirAll(tmpDir, 0o755)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
	})

	for _, dir := range nested {
		err := os.MkdirAll(filepath.Join(tmpDir, filepath.FromSlash(dir)), 0o755)
		if err != nil && !os.IsExist(err) {
			t.Fatal(err)
		}
	}

	return lsp.FileHandlerFromDirPath(tmpDir)
}

func InitPluginCache(t *testing.T, dir string) {
	pluginCacheDir := filepath.Join(dir, ".terraform", "plugins")
	err := os.MkdirAll(pluginCacheDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(pluginCacheDir, "selections.json"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
}
