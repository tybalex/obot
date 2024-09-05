package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/storage"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Request struct {
	http.ResponseWriter
	*http.Request
	GPTClient *gptscript.GPTScript
	Storage   storage.Client
	User      user.Info
}

func (r *Request) Body() ([]byte, error) {
	return io.ReadAll(io.LimitReader(r.Request.Body, 1<<20))
}

func (r *Request) JSON(obj any) error {
	r.ResponseWriter.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(r.ResponseWriter).Encode(obj)
}

func (r *Request) Write(data []byte) (int, error) {
	return r.ResponseWriter.Write(data)
}

func (r *Request) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *Request) List(obj client.ObjectList) error {
	namespace := r.Namespace()
	return r.Storage.List(r.Request.Context(), obj, &client.ListOptions{
		Namespace: namespace,
	})
}

func (r *Request) Delete(obj client.Object) error {
	return r.Storage.Delete(r.Request.Context(), obj)
}

func (r *Request) Get(obj client.Object, name string) error {
	namespace := r.Namespace()
	err := r.Storage.Get(r.Request.Context(), router.Key(namespace, name), obj)
	if apierrors.IsNotFound(err) {
		gvk, _ := r.Storage.GroupVersionKindFor(obj)
		return NewErrHttp(http.StatusNotFound, fmt.Sprintf("%s %s not found", strings.ToLower(gvk.Kind), name))
	}
	return err
}

func (r *Request) Create(obj client.Object) error {
	return r.Storage.Create(r.Request.Context(), obj)
}

func (r *Request) Update(obj client.Object) error {
	return r.Storage.Update(r.Request.Context(), obj)
}

func (r *Request) Namespace() string {
	if r.User.GetUID() != "" {
		return r.User.GetUID()
	}
	if r.User.GetName() != "" {
		return r.User.GetName()
	}
	return "default"
}
