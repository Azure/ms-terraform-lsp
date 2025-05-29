package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/ms-terraform-lsp/internal/langserver"
	"github.com/Azure/ms-terraform-lsp/internal/langserver/session"
	"github.com/Azure/ms-terraform-lsp/internal/protocol"
)

func TestCodeAction_withoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "textDocument/codeAction",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "terraform",
			"text": "provider \"github\" {}",
			"uri": "%s/main.tf"
		}
	}`, TempDir(t).URI()),
	}, session.SessionNotInitialized.Err())
}

func TestCodeAction_permission(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Dir())

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	config, err := os.ReadFile(fmt.Sprintf("./testdata/%s/main.tf", t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	reqParams := buildReqParamsCodeAction(1, 1, 13, 1, tmpDir.URI())

	expected := []protocol.CodeAction{
		{
			Title: "Generate Custom Role",
			Kind:  "refactor.rewrite",
			Edit:  protocol.WorkspaceEdit{},
			Command: &protocol.Command{
				Title:   "Generate Custom Role",
				Command: "ms-terraform.aztfauthorize",
				Arguments: []json.RawMessage{
					[]byte(reqParams),
					[]byte(`{"generateForMissingPermission":false}`),
				},
			},
		},
		{
			Title: "Generate Custom Role for Missing Permissions",
			Kind:  "refactor.rewrite",
			Edit:  protocol.WorkspaceEdit{},
			Command: &protocol.Command{
				Title:   "Generate Custom Role for Missing Permissions",
				Command: "ms-terraform.aztfauthorize",
				Arguments: []json.RawMessage{
					[]byte(reqParams),
					[]byte(`{"generateForMissingPermission":true}`),
				},
			},
		},
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
		Method:    "textDocument/codeAction",
		ReqParams: reqParams,
	})

	var result []protocol.CodeAction
	err = json.Unmarshal(rawResponse.Result, &result)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func buildReqParamsCodeAction(startLine, startCharacter, endLine, endCharacter int, uri string) string {
	param := protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: protocol.DocumentURI(fmt.Sprintf("%s/main.tf", uri)),
		},
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(startLine - 1),
				Character: uint32(startCharacter - 1),
			},
			End: protocol.Position{
				Line:      uint32(endLine - 1),
				Character: uint32(endCharacter - 1),
			},
		},
		Context: protocol.CodeActionContext{},
	}
	paramJson, _ := json.Marshal(param)
	return string(paramJson)
}
