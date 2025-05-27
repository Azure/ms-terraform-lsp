package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azurerm-lsp/internal/langserver"
	"github.com/Azure/azurerm-lsp/internal/langserver/session"
	"github.com/Azure/azurerm-lsp/internal/protocol"
)

func TestCompletion_withoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "textDocument/completion",
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

func TestCompletion_templates(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}

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
	ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/didOpen",
		ReqParams: buildReqParamsTextDocument(string(config), tmpDir.URI()),
	})

	rawResponse := ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/completion",
		ReqParams: buildReqParamsCompletion(1, 9, tmpDir.URI()),
	})

	var result protocol.CompletionList
	err = json.Unmarshal(rawResponse.Result, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(result.Items) < 100 {
		t.Fatalf("expected at least 100 items, got %d", len(result.Items))
	}

	rawResponse = ls.Call(t, &langserver.CallRequest{
		Method:    "textDocument/completion",
		ReqParams: buildReqParamsCompletion(8, 9, tmpDir.URI()),
	})

	err = json.Unmarshal(rawResponse.Result, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(result.Items) < 100 {
		t.Fatalf("expected at least 100 items, got %d", len(result.Items))
	}
}

func TestCompletion_properties(t *testing.T) {
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
	}`, tmpDir.URI()),
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
		Method:    "textDocument/completion",
		ReqParams: buildReqParamsCompletion(3, 3, tmpDir.URI()),
	}, string(expectRaw))
}

func TestCompletion_propertyValues(t *testing.T) {
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
	}`, tmpDir.URI()),
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
		Method:    "textDocument/completion",
		ReqParams: buildReqParamsCompletion(5, 25, tmpDir.URI()),
	}, string(expectRaw))
}

func buildReqParamsCompletion(line int, character int, uri string) string {
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

func buildReqParamsTextDocument(text string, uri string) string {
	param := make(map[string]interface{})
	textDocument := make(map[string]interface{})
	textDocument["version"] = 0
	textDocument["languageId"] = "terraform"
	textDocument["text"] = text
	textDocument["uri"] = fmt.Sprintf("%s/main.tf", uri)
	param["textDocument"] = textDocument
	paramJson, _ := json.Marshal(param)
	return string(paramJson)
}
