package expression

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/dop251/goja"
	"github.com/gptscript-ai/otto/pkg/gz"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type proxied interface {
	Keys() ([]string, error)
	Get(string) (any, bool, error)
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
	if len(step.Spec.ForItem) == 0 && step.Spec.ParentWorkflowStepName != "" {
		var parentStep v1.WorkflowStep
		if err := client.Get(ctx, router.Key(step.Namespace, step.Spec.ParentWorkflowStepName), &parentStep); err != nil {
			return err
		}
		return setForItem(ctx, client, vm, &parentStep)
	}

	if len(step.Spec.ForItem) == 0 || step.Spec.Step.ForEach == nil {
		return nil
	}

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

func Eval(ctx context.Context, client kclient.Client, step *v1.WorkflowStep, expr string) (any, error) {
	var workflowExecution v1.WorkflowExecution
	if err := client.Get(ctx, router.Key(step.Namespace, step.Spec.WorkflowExecutionName), &workflowExecution); err != nil {
		return nil, err
	}

	var (
		steps = &Steps{
			step:   step,
			ctx:    ctx,
			client: client,
		}
		workflow = &Workflow{
			execution: &workflowExecution,
			steps:     steps,
		}
	)

	vm := goja.New()
	if err := setProxied(vm, "workflow", workflow); err != nil {
		return nil, err
	}

	if err := setProxied(vm, "steps", steps); err != nil {
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

func setProxied(vm *goja.Runtime, name string, proxied proxied) error {
	return vm.Set(name, newWrapper(vm, proxied))
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
