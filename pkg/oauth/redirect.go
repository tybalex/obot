package oauth

import (
	"net/http"
)

func IsOAuthCallbackResponse(r *http.Request) bool {
	return r.URL.Path == "/" &&
		(r.URL.Query().Get("code") != "" ||
			r.URL.Query().Get("error") != "" ||
			r.URL.Query().Get("state") != "")
}

func HandleOAuthRedirect(w http.ResponseWriter, r *http.Request) bool {
	if !IsOAuthCallbackResponse(r) {
		return false
	}
	redirectURL := r.URL
	redirectURL.Path = "/oauth2/callback"
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
	return true
}
