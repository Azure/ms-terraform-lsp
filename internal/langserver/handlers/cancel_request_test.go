package handlers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/Azure/ms-terraform-lsp/internal/langserver"
	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/handler"
)

func TestLangServer_cancelRequest(t *testing.T) {
	tmpDir := TempDir(t)

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
		AdditionalHandlers: map[string]handler.Func{
			"$/sleepTicker": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
				ticker := time.NewTicker(100 * time.Millisecond)

				ctx, cancelFunc := context.WithTimeout(ctx, 1*time.Second)
				t.Cleanup(cancelFunc)

				var wg sync.WaitGroup
				wg.Add(1)
				go func(ctx context.Context) {
					defer wg.Done()
					for {
						select {
						case <-ctx.Done():
							ticker.Stop()
							return
						case <-ticker.C:
							log.Printf("tick at %s", time.Now())
						}
					}
				}(ctx)
				wg.Wait()

				return nil, ctx.Err()
			},
		},
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ls.CallAndExpectError(t, &langserver.CallRequest{
			Method:    "$/sleepTicker",
			ReqParams: `{}`,
		}, context.Canceled)
	}()
	time.Sleep(100 * time.Millisecond)
	ls.Call(t, &langserver.CallRequest{
		Method:    "$/cancelRequest",
		ReqParams: fmt.Sprintf(`{"id": %d}`, 2),
	})
	wg.Wait()
}
