package expression

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/dop251/goja"
	"github.com/gptscript-ai/otto/pkg/gz"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Steps struct {
	step   *v1.WorkflowStep
	ctx    context.Context
	client kclient.Client
}

func (s *Steps) Keys() (result []string, _ error) {
	keySet := map[string]struct{}{}
	if s.step.Spec.ParentWorkflowStepName != "" {
		var parent v1.WorkflowStep
		if err := s.client.Get(s.ctx, router.Key(s.step.Namespace, s.step.Spec.ParentWorkflowStepName), &parent); err != nil {
			return nil, err
		}
		parentStep := &Steps{
			step:   &parent,
			ctx:    s.ctx,
			client: s.client,
		}
		parentKeys, err := parentStep.Keys()
		if err != nil {
			return nil, err
		}
		for _, key := range parentKeys {
			keySet[key] = struct{}{}
		}
	}

	sibs, err := s.getSiblings()
	if err != nil {
		return nil, err
	}
	for _, item := range sibs {
		keySet[item.Spec.Step.Name] = struct{}{}
	}

	return slices.Sorted(maps.Keys(keySet)), nil
}

func (s *Steps) Get(key string) (any, bool, error) {
	return s.GetStep(key)
}

func (s *Steps) getSiblings() (result []v1.WorkflowStep, _ error) {
	var steps v1.WorkflowStepList
	if err := s.client.List(s.ctx, &steps, kclient.InNamespace(s.step.Namespace)); err != nil {
		return nil, err
	}
	for _, item := range steps.Items {
		if item.Name == s.step.Name {
			continue
		}
		if item.Spec.ParentWorkflowStepName == s.step.Spec.ParentWorkflowStepName {
			result = append(result, item)
		}
	}
	sort.Slice(steps.Items, func(i, j int) bool {
		return steps.Items[i].Name < steps.Items[j].Name
	})
	return
}

func (s *Steps) GetStep(name string) (*Step, bool, error) {
	if name == "parent" {
		var parent v1.WorkflowStep
		if err := s.client.Get(s.ctx, router.Key(s.step.Namespace, s.step.Spec.ParentWorkflowStepName), &parent); apierrors.IsNotFound(err) {
			return nil, false, nil
		} else if err != nil {
			return nil, false, err
		}
		return &Step{step: &parent, ctx: s.ctx, client: s.client}, true, nil
	}

	siblings, err := s.getSiblings()
	if err != nil {
		return nil, false, err
	}

	for _, item := range siblings {
		if strings.EqualFold(item.Spec.Step.Name, name) {
			return &Step{step: &item, ctx: s.ctx, client: s.client}, true, nil
		}
	}

	return nil, false, nil
}

type proxied interface {
	Keys() ([]string, error)
	Get(string) (any, bool, error)
}

type Step struct {
	step   *v1.WorkflowStep
	ctx    context.Context
	client kclient.Client
}

func (s *Step) Keys() ([]string, error) {
	return []string{
		"input",
		"output",
	}, nil
}

func (s *Step) Get(key string) (any, bool, error) {
	var run v1.Run
	if err := s.client.Get(s.ctx, router.Key(s.step.Namespace, s.step.Status.LastRunName), &run); err != nil {
		return nil, false, err
	}

	switch strings.ToLower(key) {
	case "input":
		return &StringWrapper{value: run.Spec.Input}, true, nil
	case "output":
		return &StringWrapper{value: run.Status.Output}, true, nil
	}
	return nil, false, nil
}

type StringWrapper struct {
	value string
}

func (s *StringWrapper) Keys() ([]string, error) {
	return []string{
		"text",
		"json",
	}, nil
}

func (s *StringWrapper) Get(key string) (any, bool, error) {
	switch strings.ToLower(key) {
	case "text":
		return s.value, true, nil
	case "json":
		var obj any
		err := json.Unmarshal([]byte(s.value), &obj)
		return obj, err == nil, err
	default:
		return nil, false, nil
	}
}

func EvalArray(ctx context.Context, client kclient.Client, step *v1.WorkflowStep, expr string) ([]any, error) {
	if expr == "" {
		return nil, nil
	}
	result, err := Eval(ctx, client, step, expr)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]any); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("while evaluating %q expected array, got %T", expr, result)
}

func EvalBool(ctx context.Context, client kclient.Client, step *v1.WorkflowStep, expr string) (bool, error) {
	if expr == "" {
		return false, nil
	}
	result, err := Eval(ctx, client, step, expr)
	if err != nil {
		return false, err
	}
	if b, ok := result.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("while evaluating %q expected boolean, got %T", expr, result)
}

func EvalString(ctx context.Context, client kclient.Client, step *v1.WorkflowStep, expr string) (string, error) {
	if expr == "" {
		return "", nil
	}
	result, err := Eval(ctx, client, step, expr)
	if err != nil {
		return "", err
	}
	if str, ok := result.(string); ok {
		return str, nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func setForItem(ctx context.Context, client kclient.Client, vm *goja.Runtime, step *v1.WorkflowStep) error {
	if step.Spec.ParentWorkflowStepName != "" {
		if err := setForItem(ctx, client, vm, step); err != nil {
			return err
		}
	}
	if len(step.Spec.ForItem) > 0 {
		var obj any
		if err := gz.Decompress(&obj, step.Spec.ForItem); err != nil {
			return err
		}
		var itemName = "item"
		if step.Spec.Step.ForEach.Var != "" {
			itemName = step.Spec.Step.ForEach.Var
		}
		return vm.Set(itemName, obj)
	}
	return nil
}

func Eval(ctx context.Context, client kclient.Client, step *v1.WorkflowStep, expr string) (any, error) {
	vm := goja.New()
	err := vm.Set("steps", newWrapper(vm, &Steps{
		step:   step,
		ctx:    ctx,
		client: client,
	}))
	if err != nil {
		return nil, err
	}

	if err := setForItem(ctx, client, vm, step); err != nil {
		return nil, err
	}

	x, err := vm.RunString(expr)
	if err != nil {
		return nil, err
	}
	return x.Export(), nil
}

type wrapper struct {
	vm    *goja.Runtime
	proxy proxied
}

func newWrapper(vm *goja.Runtime, proxied proxied) goja.Proxy {
	wrapper := &wrapper{
		vm:    vm,
		proxy: proxied,
	}
	return vm.NewProxy(vm.NewObject(), &goja.ProxyTrapConfig{
		Get:     wrapper.Get,
		OwnKeys: wrapper.OwnKeys,
		Has:     wrapper.Has,
	})
}

func (w *wrapper) toValue(value any) goja.Value {
	switch v := value.(type) {
	case proxied:
		return w.vm.ToValue(newWrapper(w.vm, v))
	default:
		return w.vm.ToValue(v)
	}
}

func (w *wrapper) Get(_ *goja.Object, property string, _ goja.Value) (value goja.Value) {
	v, found, err := w.proxy.Get(property)
	if err != nil {
		return w.vm.NewGoError(err)
	} else if !found {
		return goja.Undefined()
	}
	return w.toValue(v)
}

func (w *wrapper) OwnKeys(_ *goja.Object) (object *goja.Object) {
	keys, err := w.proxy.Keys()
	if err != nil {
		return w.vm.NewGoError(err)
	}
	var objs []any
	for _, key := range keys {
		objs = append(objs, key)
	}
	return w.vm.NewArray(objs...)
}

func (w *wrapper) Has(_ *goja.Object, property string) (available bool) {
	keys, err := w.proxy.Keys()
	if err != nil {
		return false
	}
	return slices.Contains(keys, property)
}
