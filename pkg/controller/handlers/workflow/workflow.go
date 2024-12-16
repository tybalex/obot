package workflow

import (
	v1 "github.com/acorn-io/acorn/pkg/storage/apis/otto.otto8.ai/v1"
	"github.com/acorn-io/nah/pkg/router"
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
