package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
)

type Component interface {
	// Render the template.
	Render(ctx context.Context, w io.Writer) error
}

func Render(rw http.ResponseWriter, req *http.Request, component Component) error {
	templ.Handler(component).ServeHTTP(rw, req)
	return nil
}
