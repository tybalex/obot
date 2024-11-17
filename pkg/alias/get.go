package alias

import (
	"context"
	"errors"
	"strings"

	"github.com/otto8-ai/nah/pkg/name"
	"github.com/otto8-ai/nah/pkg/router"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.otto8.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Get(ctx context.Context, c kclient.Client, obj v1.Aliasable, namespace string, name string) error {
	var errLookup error
	if namespace == "" {
		gvk, err := c.GroupVersionKindFor(obj.(kclient.Object))
		if err != nil {
			return err
		}
		errLookup = apierrors.NewNotFound(schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind,
		}, name)
	} else {
		errLookup = c.Get(ctx, router.Key(namespace, name), obj.(kclient.Object))
		if kclient.IgnoreNotFound(errLookup) != nil {
			return errLookup
		} else if errLookup == nil {
			return nil
		}
	}

	gvk, err := c.GroupVersionKindFor(obj.(kclient.Object))
	if err != nil {
		return err
	}

	var alias v1.Alias
	if err := c.Get(ctx, router.Key("", Key(gvk, obj, name)), &alias); apierrors.IsNotFound(err) {
		return errLookup
	} else if err != nil {
		return errors.Join(errLookup, err)
	} else if alias.Spec.TargetKind != gvk.Kind {
		return errLookup
	}

	return c.Get(ctx, router.Key(alias.Spec.TargetNamespace, alias.Spec.TargetName), obj.(kclient.Object))
}

func keyFromName(scope, id string) string {
	return strings.ToLower(name.SafeHashConcatName(id, scope))
}

func getScope(gvk schema.GroupVersionKind, obj v1.Aliasable) string {
	if scoped, ok := obj.(v1.AliasScoped); ok && scoped.GetAliasScope() != "" {
		return scoped.GetAliasScope()
	}

	return gvk.Kind
}

func Key(gvk schema.GroupVersionKind, obj v1.Aliasable, id string) string {
	return keyFromName(getScope(gvk, obj), id)
}
