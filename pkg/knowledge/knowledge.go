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

func (i *Ingester) IngestKnowledge(ctx context.Context, agentName, namespace, knowledgeWorkspaceID string) (*invoke.Response, error) {
	return i.invoker.SystemAction(
		ctx,
		"ingest-",
		agentName,
		namespace,
		fullKnowledgeTool(i.knowledgeTool, "ingest.gpt"),
		workspace.GetDir(knowledgeWorkspaceID),
		// Below are environment variables used by the ingest tool
		"GPTSCRIPT_DATASET="+workspace.KnowledgeIDFromWorkspaceID(knowledgeWorkspaceID),
		"KNOW_JSON=true",
	)
}

func (i *Ingester) DeleteKnowledge(ctx context.Context, agentName, namespace, knowledgeWorkspaceID string) (*invoke.Response, error) {
	return i.invoker.SystemAction(
		ctx,
		"ingest-delete-",
		agentName,
		namespace,
		fullKnowledgeTool(i.knowledgeTool, "delete.gpt"),
		workspace.KnowledgeIDFromWorkspaceID(knowledgeWorkspaceID),
	)
}

func fullKnowledgeTool(knowledgeTool, subTool string) string {
	knowledgeTool, tag, ok := strings.Cut(knowledgeTool, "@")
	if ok {
		tag = "@" + tag
	}

	return fmt.Sprintf("%s/%s%s", knowledgeTool, subTool, tag)
}
