package api

import (
	"errors"
	"fmt"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type ErrHTTP struct {
	Code    int
	Message string
}

func (e *ErrHTTP) Error() string {
	return fmt.Sprintf("error code %d (%s): %s", e.Code, http.StatusText(e.Code), e.Message)
}

func IsHTTPCode(err error, code int) bool {
	if err == nil {
		return false
	}
	if errHttp := (*ErrHTTP)(nil); errors.As(err, &errHttp) {
		return errHttp.Code == code
	} else if errMeta := (*apierrors.StatusError)(nil); errors.As(err, &errMeta) {
		return errMeta.ErrStatus.Code == int32(code)
	}
	return false
}

func NewErrHttp(code int, message string) *ErrHTTP {
	return &ErrHTTP{
		Code:    code,
		Message: message,
	}
}

func IsConflict(err error) bool {
	return IsHTTPCode(err, http.StatusConflict)
}

func NewErrBadRequest(message string, args ...interface{}) *ErrHTTP {
	return NewErrHttp(http.StatusBadRequest, fmt.Sprintf(message, args...))
}

var ErrMustAuth = &ErrHTTP{
	Code:    http.StatusUnauthorized,
	Message: "unauthorized request, must authenticate",
}
