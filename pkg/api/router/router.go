package router

import (
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api/handlers"
	"github.com/gptscript-ai/otto/pkg/services"
)

func Router(services *services.Services) (http.Handler, error) {
	w := services.APIServer.Wrap
	mux := http.NewServeMux()

	agents := handlers.AgentHandler{
		WorkspaceClient:   services.WorkspaceClient,
		WorkspaceProvider: "directory",
	}
	invoker := handlers.InvokeHandler{Invoker: services.Invoker}
	threads := handlers.ThreadHandler{}
	runs := handlers.RunHandler{}

	// Agents
	mux.Handle("GET /agents", w(agents.List))
	mux.Handle("POST /agents", w(agents.Create))
	mux.Handle("PUT /agents/{id}", w(agents.Update))
	mux.Handle("DELETE /agents/{id}", w(agents.Delete))

	// Invoker
	mux.Handle("POST /invoke/{agent}", w(invoker.Invoke))
	mux.Handle("POST /invoke/{agent}/threads/{thread}", w(invoker.Invoke))

	// Threads
	mux.Handle("GET /threads", w(threads.List))
	mux.Handle("GET /agents/{agent}/threads", w(threads.List))

	// Runs
	mux.Handle("GET /runs", w(runs.List))
	mux.Handle("GET /runs/{run}/debug", w(runs.Debug))
	mux.Handle("GET /threads/{thread}/runs", w(runs.List))
	mux.Handle("GET /agents/{agent}/runs", w(runs.List))
	mux.Handle("GET /agents/{agent}/threads/{thread}/runs", w(runs.List))

	return mux, nil
}
