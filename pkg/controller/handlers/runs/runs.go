package runs

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DeleteRunState(req router.Request, resp router.Response) error {
	run := req.Object.(*v1.Run)
	return req.Delete(&v1.RunState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      run.Name,
			Namespace: run.Namespace,
		},
	})
}

func Cleanup(req router.Request, resp router.Response) error {
	run := req.Object.(*v1.Run)
	var thread v1.Thread

	if err := req.Get(&thread, run.Namespace, run.Spec.ThreadName); apierrors.IsNotFound(err) {
		return req.Delete(run)
	} else if err != nil {
		return err
	}

	return nil
}
