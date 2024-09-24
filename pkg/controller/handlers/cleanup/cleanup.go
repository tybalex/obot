package cleanup

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/acorn-io/baaah/pkg/uncached"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type refs interface {
	DeleteRefs() []v1.Ref
}

func Cleanup(req router.Request, resp router.Response) error {
	toDelete := req.Object.(refs)

	for _, ref := range toDelete.DeleteRefs() {
		if ref.Name == "" {
			continue
		}
		if err := req.Get(ref.ObjType, req.Namespace, ref.Name); apierrors.IsNotFound(err) {
			if err := req.Get(uncached.Get(ref.ObjType), req.Namespace, ref.Name); apierrors.IsNotFound(err) {
				return req.Delete(req.Object)
			}
		}
	}

	return nil
}
