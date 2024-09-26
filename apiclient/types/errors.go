package types

import (
	"fmt"
	"net/http"
)

type ErrHTTP struct {
	Code    int
	Message string
}

func (e *ErrHTTP) Error() string {
	return fmt.Sprintf("error code %d (%s): %s", e.Code, http.StatusText(e.Code), e.Message)
}

func NewErrHttp(code int, message string) *ErrHTTP {
	return &ErrHTTP{
		Code:    code,
		Message: message,
	}
}
