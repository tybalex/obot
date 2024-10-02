package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gptscript-ai/otto/apiclient"
	"github.com/gptscript-ai/otto/ui/pages"
	"github.com/gptscript-ai/otto/ui/webcontext"
)

func Thread(rw http.ResponseWriter, req *http.Request) error {
	var (
		threadID = req.PathValue("id")
		ctx      = req.Context()
		c        = webcontext.Client(ctx)
	)

	thread, err := c.GetThread(ctx, threadID)
	if err != nil {
		return err
	}

	return Render(rw, req, pages.Thread(pages.ThreadData{
		Thread: *thread,
	}))
}

func ThreadEvents(rw http.ResponseWriter, req *http.Request) error {
	var (
		threadID = req.PathValue("id")
		ctx      = req.Context()
		c        = webcontext.Client(ctx)
	)

	result, err := c.ThreadEvents(ctx, threadID, apiclient.ThreadEventsOptions{})
	if err != nil {
		return err
	}

	rw.Header().Set("Content-Type", "text/event-stream")

	if err := writeDataEvent(rw, map[string]any{}, "start", ""); err != nil {
		return err
	}

	for event := range result {
		if err := writeDataEvent(rw, event, "", ""); err != nil {
			return err
		}
		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}
	}

	if err := writeDataEvent(rw, map[string]any{}, "close", ""); err != nil {
		return err
	}

	return nil
}

func writeDataEvent(r io.Writer, obj any, event, id string) error {
	buf := &bytes.Buffer{}
	if event != "" {
		buf.WriteString("event: ")
		buf.WriteString(event)
		buf.WriteString("\n")
	}
	if id != "" {
		buf.WriteString("id: ")
		buf.WriteString(id)
		buf.WriteString("\n")
	}

	buf.WriteString("data: ")
	if err := json.NewEncoder(buf).Encode(obj); err != nil {
		return err
	}
	buf.WriteString("\n\n")
	_, err := buf.WriteTo(r)
	return err
}
