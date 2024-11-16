package alias

import (
	"github.com/otto8-ai/nah/pkg/router"
	"github.com/otto8-ai/otto8/pkg/alias"
	"github.com/otto8-ai/otto8/pkg/create"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.otto8.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func matches(alias *v1.Alias, obj kclient.Object) bool {
	return alias.Spec.TargetName == obj.GetName() &&
		alias.Spec.TargetNamespace == obj.GetNamespace() &&
		alias.Spec.TargetKind == obj.GetObjectKind().GroupVersionKind().Kind
}

func AssignAlias(req router.Request, _ router.Response) error {
	aliasable := req.Object.(v1.Aliasable)

	if aliasable.GetAliasName() == "" {
		return nil
	}

	gvk, err := req.Client.GroupVersionKindFor(req.Object)
	if err != nil {
		return err
	}

	key := alias.Key(gvk, aliasable, aliasable.GetAliasName())
	alias := &v1.Alias{
		ObjectMeta: metav1.ObjectMeta{
			Name: key,
		},
		Spec: v1.AliasSpec{
			Name:            aliasable.GetAliasName(),
			TargetName:      req.Object.GetName(),
			TargetNamespace: req.Object.GetNamespace(),
			TargetKind:      gvk.Kind,
		},
		Status: v1.EmptyStatus{},
	}
	if err := create.IfNotExists(req.Ctx, req.Client, alias); err != nil {
		return err
	}

	if !matches(alias, req.Object) {
		return nil
	}

	if !aliasable.IsAssigned() {
		aliasable.SetAssigned()
		return req.Client.Status().Update(req.Ctx, req.Object)
	}

	return nil
}
