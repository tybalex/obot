package workspace

import (
	"strings"
)

func GetDir(workspaceID string) string {
	_, path, _ := strings.Cut(workspaceID, "://")
	return path
}

func KnowledgeIDFromWorkspaceID(workspaceID string) string {
	return strings.ReplaceAll(workspaceID, " ", "_")
}
