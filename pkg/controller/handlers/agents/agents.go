package agents

import (
	"context"

	"github.com/acorn-io/baaah/pkg/name"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/otto8-ai/otto8/pkg/create"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.otto8.ai/v1"
	"github.com/otto8-ai/otto8/pkg/system"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func createWorkspace(ctx context.Context, c kclient.Client, agent *v1.Agent) error {
	if agent.Status.WorkspaceName != "" {
		return nil
	}

	ws := &v1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  agent.Namespace,
			Name:       name.SafeConcatName(system.WorkspacePrefix, agent.Name),
			Finalizers: []string{v1.WorkspaceFinalizer},
		},
		Spec: v1.WorkspaceSpec{
			AgentName: agent.Name,
		},
	}
	if err := create.OrGet(ctx, c, ws); err != nil {
		return err
	}

	agent.Status.WorkspaceName = ws.Name
	return c.Status().Update(ctx, agent)
}

func createKnowledgeSet(ctx context.Context, c kclient.Client, agent *v1.Agent) error {
	if len(agent.Status.KnowledgeSetNames) > 0 {
		return nil
	}

	ks := &v1.KnowledgeSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  agent.Namespace,
			Name:       name.SafeConcatName(system.KnowledgeSetPrefix, agent.Name),
			Finalizers: []string{v1.KnowledgeSetFinalizer},
		},
		Spec: v1.KnowledgeSetSpec{
			AgentName: agent.Name,
		},
	}
	if err := create.OrGet(ctx, c, ks); err != nil {
		return err
	}

	agent.Status.KnowledgeSetNames = append(agent.Status.KnowledgeSetNames, ks.Name)
	return c.Status().Update(ctx, agent)
}

func CreateWorkspaceAndKnowledgeSet(req router.Request, _ router.Response) error {
	agent := req.Object.(*v1.Agent)

	if err := createWorkspace(req.Ctx, req.Client, agent); err != nil {
		return err
	}

	return createKnowledgeSet(req.Ctx, req.Client, agent)
}
