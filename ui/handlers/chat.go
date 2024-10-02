package handlers

import (
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/gptscript-ai/otto/apiclient"
	"github.com/gptscript-ai/otto/apiclient/types"
	"github.com/gptscript-ai/otto/ui/components"
	"github.com/gptscript-ai/otto/ui/pages"
	"github.com/gptscript-ai/otto/ui/webcontext"
)

func Chat(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return Render(w, r, pages.Chat())
	}

	var (
		ctx      = r.Context()
		c        = webcontext.Client(ctx)
		agentID  = "otto"
		threadID = "t1-user"
	)

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		return err
	}

	if !utf8.Valid(body) {
		return &types.ErrHTTP{
			Code:    http.StatusBadRequest,
			Message: "Invalid UTF-8 in request body",
		}
	}

	_, err = c.Invoke(ctx, agentID, string(body), apiclient.InvokeOptions{
		ThreadID: threadID,
		Async:    true,
	})
	return Render(w, r, components.ChatResponse(err))
}
