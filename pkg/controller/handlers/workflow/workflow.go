package workflow

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func EnsureIDs(req router.Request, resp router.Response) error {
	wf := req.Object.(*v1.Workflow)
	manifestWithIDS := PopulateIDs(wf.Spec.Manifest)
	if !equality.Semantic.DeepEqual(wf.Spec.Manifest, manifestWithIDS) {
		wf.Spec.Manifest = manifestWithIDS
		return req.Client.Update(req.Ctx, wf)
	}
	return nil
}
