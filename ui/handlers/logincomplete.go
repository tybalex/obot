package handlers

import (
	"net/http"

	"github.com/otto8-ai/otto8/ui/pages"
)

func LoginComplete(rw http.ResponseWriter, req *http.Request) error {
	return Render(rw, req, pages.LoginComplete())
}
