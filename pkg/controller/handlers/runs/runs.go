package runs

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/mvl"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = mvl.Package()

type Handler struct {
	invoker *invoke.Invoker
}

func New(invoker *invoke.Invoker) *Handler {
	return &Handler{invoker: invoker}
}

func (*Handler) DeleteRunState(req router.Request, resp router.Response) error {
	run := req.Object.(*v1.Run)
	return client.IgnoreNotFound(req.Delete(&v1.RunState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      run.Name,
			Namespace: run.Namespace,
		},
	}))
}

func (h *Handler) Resume(req router.Request, resp router.Response) error {
	run := req.Object.(*v1.Run)
	var thread v1.Thread

	if !run.Spec.Background || run.Status.State.IsTerminal() || run.Status.State == gptscript.Continue {
		return nil
	}

	if err := req.Get(&thread, run.Namespace, run.Spec.ThreadName); apierrors.IsNotFound(err) {
		return req.Delete(run)
	} else if err != nil {
		return err
	}

	return h.invoker.Resume(req.Ctx, req.Client, &thread, run)
}
