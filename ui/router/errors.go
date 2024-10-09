package router

import (
	gerrors "errors"
	"net/http"

	"github.com/otto8-ai/otto8/apiclient/types"
)

func errors(f func(rw http.ResponseWriter, req *http.Request) error) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if err := f(rw, req); err != nil {
			if httpErr := (*types.ErrHTTP)(nil); gerrors.As(err, &httpErr) {
				http.Error(rw, httpErr.Message, httpErr.Code)
			} else {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
