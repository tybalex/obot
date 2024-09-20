package knowledge

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/otto/pkg/invoke"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/workspace"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Knowledgeable interface {
	client.Object
	GetKnowledgeWorkspaceStatus() *v1.KnowledgeWorkspaceStatus
}

type Ingester struct {
	invoker       *invoke.Invoker
	knowledgeTool string
}

func NewIngester(invoker *invoke.Invoker, knowledgeTool string) *Ingester {
	return &Ingester{
		invoker:       invoker,
		knowledgeTool: knowledgeTool,
	}
}

func (i *Ingester) IngestKnowledge(ctx context.Context, agentName, namespace, knowledgeWorkspaceID string) error {
	knowledgeTool, tag, ok := strings.Cut(i.knowledgeTool, "@")
	if ok {
		tag = "@" + tag
	}

	run, err := i.invoker.SystemAction(
		ctx,
		"ingest-",
		agentName,
		namespace,
		knowledgeTool+"/ingest.gpt"+tag,
		workspace.GetDir(knowledgeWorkspaceID),
		// These are environment variables passed to the script
		"GPTSCRIPT_DATASET="+workspace.KnowledgeIDFromWorkspaceID(knowledgeWorkspaceID),
	)
	if err != nil {
		return err
	}

	run.Wait()
	if run.Run.Status.Error != "" {
		return fmt.Errorf("failed to ingest knowledge: %s", run.Run.Status.Error)
	}
	return nil
}

func (i *Ingester) DeleteKnowledge(ctx context.Context, agentName, namespace, knowledgeWorkspaceID string) error {
	knowledgeTool, tag, ok := strings.Cut(i.knowledgeTool, "@")
	if ok {
		tag = "@" + tag
	}

	run, err := i.invoker.SystemAction(
		ctx,
		"ingest-delete-",
		agentName,
		namespace,
		knowledgeTool+"/delete.gpt"+tag,
		workspace.KnowledgeIDFromWorkspaceID(knowledgeWorkspaceID),
	)
	if err != nil {
		return err
	}

	run.Wait()
	if run.Run.Status.Error != "" {
		return fmt.Errorf("failed to delete knowledge: %s", run.Run.Status.Error)
	}
	return nil
}
