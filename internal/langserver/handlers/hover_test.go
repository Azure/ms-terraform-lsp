package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/ms-terraform-lsp/internal/langserver"
	"github.com/Azure/ms-terraform-lsp/internal/langserver/session"
)

func TestHover_withoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "textDocument/hover",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			},
			"position": {
				"character": 0,
				"line": 1
			}
		}`, TempDir(t).URI()),
	}, session.SessionNotInitialized.Err())
}

func TestHover_property(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	expectRaw, err := os.ReadFile(fmt.Sprintf("./testdata/%s/expect.json", t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	expectRaw = []byte(strings.ReplaceAll(string(expectRaw), "<", "\\u003c"))
	expectRaw = []byte(strings.ReplaceAll(string(expectRaw), ">", "\\u003e"))
	expectRaw = []byte(strings.ReplaceAll(string(expectRaw), "&", "\\u0026"))

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, TempDir(t).URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method:    "textDocument/hover",
		ReqParams: buildReqParamsHover(5, 8, tmpDir.URI()),
	}, string(expectRaw))
}

func TestHover_propertyValue(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	expectRaw, err := os.ReadFile(fmt.Sprintf("./testdata/%s/expect.json", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, TempDir(t).URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method:    "textDocument/hover",
		ReqParams: buildReqParamsHover(5, 16, tmpDir.URI()),
	}, string(expectRaw))
}

func TestHoverMSGraph_urlUsingExpression(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	expectRaw, err := os.ReadFile(fmt.Sprintf("./testdata/%s/expect.json", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, TempDir(t).URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method:    "textDocument/hover",
		ReqParams: buildReqParamsHover(5, 12, tmpDir.URI()),
	}, string(expectRaw))
}

func TestHoverMSGraph_prop(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	expectRaw, err := os.ReadFile(fmt.Sprintf("./testdata/%s/expect.json", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, TempDir(t).URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method:    "textDocument/hover",
		ReqParams: buildReqParamsHover(5, 11, tmpDir.URI()),
	}, string(expectRaw))
}

func TestHoverMSGraph_propInArray(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	expectRaw, err := os.ReadFile(fmt.Sprintf("./testdata/%s/expect.json", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, TempDir(t).URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method:    "textDocument/hover",
		ReqParams: buildReqParamsHover(6, 13, tmpDir.URI()),
	}, string(expectRaw))
}

func TestHoverMSGraph_resourceTitle(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	expectRaw, err := os.ReadFile(fmt.Sprintf("./testdata/%s/expect.json", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, TempDir(t).URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method:    "textDocument/hover",
		ReqParams: buildReqParamsHover(1, 23, tmpDir.URI()),
	}, string(expectRaw))
}

func TestHover_resourceTitle(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

	expectRaw, err := os.ReadFile(fmt.Sprintf("./testdata/%s/expect.json", t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	expectRaw = []byte(strings.ReplaceAll(string(expectRaw), "<", "\\u003c"))
	expectRaw = []byte(strings.ReplaceAll(string(expectRaw), ">", "\\u003e"))
	expectRaw = []byte(strings.ReplaceAll(string(expectRaw), "&", "\\u0026"))

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, TempDir(t).URI()),
	})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method:    "textDocument/hover",
		ReqParams: buildReqParamsHover(1, 6, tmpDir.URI()),
	}, string(expectRaw))
}

func buildReqParamsHover(line int, character int, uri string) string {
	param := make(map[string]interface{})
	textDocument := make(map[string]interface{})
	textDocument["uri"] = fmt.Sprintf("%s/main.tf", uri)
	param["textDocument"] = textDocument
	position := make(map[string]interface{})
	position["character"] = character - 1
	position["line"] = line - 1
	param["position"] = position
	paramJson, _ := json.Marshal(param)
	return string(paramJson)
}
