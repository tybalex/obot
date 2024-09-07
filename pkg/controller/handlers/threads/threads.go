package threads

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func Cleanup(req router.Request, resp router.Response) error {
	run := req.Object.(*v1.Thread)
	var agent v1.Agent

	if err := req.Get(&agent, run.Namespace, run.Spec.AgentName); apierrors.IsNotFound(err) {
		return req.Client.Delete(req.Ctx, run)
	} else if err != nil {
		return err
	}

	return nil
}
