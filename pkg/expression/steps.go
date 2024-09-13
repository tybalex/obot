package expression

import (
	"context"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
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

	if s.step.Spec.ParentWorkflowStepName != "" {
		var parent v1.WorkflowStep
		if err := s.client.Get(s.ctx, router.Key(s.step.Namespace, s.step.Spec.ParentWorkflowStepName), &parent); err != nil {
			return nil, false, err
		}
		parentSteps := &Steps{
			step:   &parent,
			ctx:    s.ctx,
			client: s.client,
		}
		return parentSteps.GetStep(name)
	}

	return nil, false, nil
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
