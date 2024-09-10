package workspace

import (
	"fmt"
	"os/exec"
	"strings"
)

func KnowledgeIDFromWorkspaceID(workspaceID string) string {
	return strings.ReplaceAll(workspaceID, " ", "_")
}

func IngestKnowledge(knowledgeBin, knowledgeWorkspaceID string) error {
	if err := exec.Command(knowledgeBin, "ingest", "--prune", "--dataset", KnowledgeIDFromWorkspaceID(knowledgeWorkspaceID), GetDir(knowledgeWorkspaceID)).Run(); err != nil {
		return fmt.Errorf("failed to ingest agent knowledge dataset: %w", err)
	}

	return nil
}
