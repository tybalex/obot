package cookie

import (
	"net/http"

	"k8s.io/apiserver/pkg/authentication/authenticator"
)

type Auth struct {
	next authenticator.Request
}

func New(next authenticator.Request) authenticator.Request {
	return &Auth{
		next: next,
	}
}

func (c *Auth) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	if c.next == nil {
		return nil, false, nil
	}

	token, ok := GetCookieToken(req)
	if ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.next.AuthenticateRequest(req)
}

func GetCookieToken(req *http.Request) (string, bool) {
	c, err := req.Cookie("A_SESS")
	if err != nil {
		return "", false
	}

	return c.Value, true
}
