package knowledge

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/workspace"
)

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

func (i *Ingester) IngestKnowledge(ctx context.Context, namespace, knowledgeWorkspaceID string) error {
	knowledgeTool, tag, ok := strings.Cut(i.knowledgeTool, "@")
	if ok {
		tag = "@" + tag
	}

	run, err := i.invoker.SystemAction(ctx, "ingest-", namespace, knowledgeTool+"/ingest.gpt"+tag, workspace.GetDir(knowledgeWorkspaceID), "GPTSCRIPT_DATASET="+workspace.KnowledgeIDFromWorkspaceID(knowledgeWorkspaceID))
	if err != nil {
		return err
	}

	run.Wait()
	if run.Run.Status.Error != "" {
		return fmt.Errorf("failed to ingest knowledge: %s", run.Run.Status.Error)
	}
	return nil
}

func (i *Ingester) DeleteKnowledge(ctx context.Context, namespace, knowledgeWorkspaceID string) error {
	knowledgeTool, tag, ok := strings.Cut(i.knowledgeTool, "@")
	if ok {
		tag = "@" + tag
	}

	run, err := i.invoker.SystemAction(ctx, "ingest-delete-", namespace, knowledgeTool+"/delete.gpt"+tag, workspace.KnowledgeIDFromWorkspaceID(knowledgeWorkspaceID))
	if err != nil {
		return err
	}

	run.Wait()
	if run.Run.Status.Error != "" {
		return fmt.Errorf("failed to delete knowledge: %s", run.Run.Status.Error)
	}
	return nil
}
