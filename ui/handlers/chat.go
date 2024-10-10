package handlers

import (
	"net/http"

	"github.com/otto8-ai/otto8/apiclient"
	"github.com/otto8-ai/otto8/ui/components"
	"github.com/otto8-ai/otto8/ui/pages"
	"github.com/otto8-ai/otto8/ui/webcontext"
)

func Events(w http.ResponseWriter, r *http.Request) error {
	var (
		ctx    = r.Context()
		c      = webcontext.Client(ctx)
		lastID = r.Header.Get("Last-Event-ID")
	)

	w.Header().Set("Content-Type", "text/event-stream")

	events, err := c.ThreadEvents(ctx, webcontext.ThreadID(ctx), apiclient.ThreadEventsOptions{
		Follow: true,
		RunID:  lastID,
	})
	if err != nil {
		return err
	}

	if lastID != "" {
		if err := writeDataEvent(w, map[string]any{}, "reconnect", ""); err != nil {
			return err
		}
	} else {
		if err := writeDataEvent(w, map[string]any{}, "connect", ""); err != nil {
			return err
		}
	}

	for event := range events {
		lastID = event.RunID
		if event.RunComplete {
			lastID += ":after"
		}
		if err := writeDataEvent(w, event, "", lastID); err != nil {
			return err
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	return nil
}

func Chat(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return Render(w, r, pages.Chat())
	}

	var (
		ctx      = r.Context()
		c        = webcontext.Client(ctx)
		agentID  = webcontext.AgentID(ctx)
		threadID = webcontext.ThreadID(ctx)
		input    string
	)

	if err := r.ParseForm(); err != nil {
		return err
	}

	input = r.FormValue("message")

	_, err := c.Invoke(ctx, agentID, input, apiclient.InvokeOptions{
		ThreadID: threadID,
		Async:    true,
	})
	return Render(w, r, components.ChatResponse(err))
}

func ChatSidebar(w http.ResponseWriter, r *http.Request) error {
	var (
		ctx      = r.Context()
		c        = webcontext.Client(ctx)
		threadID = webcontext.ThreadID(ctx)
	)

	files, err := c.ListFiles(ctx, apiclient.ListFileOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return err
	}
	return Render(w, r, components.NewChatSidebar(components.ChatSidebarData{
		Files: files.Items,
	}))
}
