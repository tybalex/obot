package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/otto8-ai/nah/pkg/name"
	"github.com/otto8-ai/otto8/apiclient/types"
	"github.com/otto8-ai/otto8/pkg/api"
	"github.com/otto8-ai/otto8/pkg/events"
	"github.com/otto8-ai/otto8/pkg/invoke"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.otto8.ai/v1"
	"github.com/otto8-ai/otto8/pkg/system"
	"k8s.io/apiserver/pkg/authentication/user"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type AssistantHandler struct {
	invoker   *invoke.Invoker
	events    *events.Emitter
	gptScript *gptscript.GPTScript
}

func NewAssistantHandler(invoker *invoke.Invoker, events *events.Emitter, gptScript *gptscript.GPTScript) *AssistantHandler {
	return &AssistantHandler{
		invoker:   invoker,
		events:    events,
		gptScript: gptScript,
	}
}

func (a *AssistantHandler) getAgent(req api.Context, id string) (*v1.Agent, error) {
	var ref v1.Reference
	if err := req.Get(&ref, id); err != nil {
		return nil, err
	}

	if ref.Spec.AgentName == "" {
		return nil, types.NewErrNotFound("assistant not found: %s", id)
	}

	var agent v1.Agent
	if err := req.Get(&agent, ref.Spec.AgentName); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (a *AssistantHandler) Invoke(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	agent, err := a.getAgent(req, id)
	if err != nil {
		return err
	}

	threadID := getUserThreadID(id, req.User)
	input, err := req.Body()
	if err != nil {
		return err
	}

	resp, err := a.invoker.Agent(req.Context(), req.Storage, agent, string(input), invoke.Options{
		ThreadName:   threadID,
		CreateThread: true,
		UserUID:      req.User.GetUID(),
		AgentRefName: agent.Name,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	req.ResponseWriter.Header().Set("X-Otto-Thread-Id", resp.Thread.Name)

	req.WriteHeader(http.StatusCreated)
	req.ResponseWriter.Header().Set("Content-Type", "application/json")
	return req.Write(map[string]string{
		"threadID": resp.Thread.Name,
	})
}

func (a *AssistantHandler) List(req api.Context) error {
	var refs v1.ReferenceList
	if err := req.List(&refs); err != nil {
		return err
	}

	assistants := types.AssistantList{
		Items: make([]types.Assistant, 0, len(refs.Items)),
	}
	for _, ref := range refs.Items {
		if ref.Spec.AgentName != "" {
			var agent v1.Agent
			if err := req.Get(&agent, ref.Spec.AgentName); kclient.IgnoreNotFound(err) != nil {
				return err
			} else if err == nil {
				assistants.Items = append(assistants.Items, convertAssistant(agent))
			}
		}
	}

	return req.Write(assistants)
}

func convertAssistant(agent v1.Agent) types.Assistant {
	var icons types.AgentIcons
	if agent.Spec.Manifest.Icons != nil {
		icons = *agent.Spec.Manifest.Icons
	}
	assistant := types.Assistant{
		Metadata:    MetadataFrom(&agent),
		Name:        agent.Spec.Manifest.Name,
		Description: agent.Spec.Manifest.Description,
		Icons:       icons,
	}
	assistant.ID = agent.Spec.Manifest.RefName
	return assistant
}

func getUserThreadID(agentID string, user user.Info) string {
	id := user.GetUID()
	if id == "" {
		id = "none"
	}
	return name.SafeConcatNameWithSeparatorAndLength(64, ".", system.ThreadPrefix,
		agentID, id)
}

func (a *AssistantHandler) Events(req api.Context) error {
	var (
		id       = req.PathValue("id")
		runID    = req.Request.Header.Get("Last-Event-ID")
		threadID = getUserThreadID(id, req.User)
	)

	_, events, err := a.events.Watch(req.Context(), req.Namespace(), events.WatchOptions{
		Follow:        true,
		History:       runID == "",
		LastRunName:   strings.TrimSuffix(runID, ":after"),
		MaxRuns:       10,
		After:         strings.HasSuffix(runID, ":after"),
		ThreadName:    threadID,
		WaitForThread: true,
	})
	if err != nil {
		return err
	}

	return req.WriteEvents(events)
}

func (a *AssistantHandler) Files(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if thread.Status.WorkspaceID == "" {
		return req.Write(types.FileList{Items: []types.File{}})
	}

	return listFileFromWorkspace(req.Context(), req, a.gptScript, gptscript.ListFilesInWorkspaceOptions{
		WorkspaceID: thread.Status.WorkspaceID,
		Prefix:      "files/",
	})
}

func (a *AssistantHandler) GetFile(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if thread.Status.WorkspaceID == "" {
		return types.NewErrNotFound("no workspace found for assistant %s", id)
	}

	return getFileInWorkspace(req.Context(), req, a.gptScript, thread.Status.WorkspaceID, "files/")
}

func (a *AssistantHandler) UploadFile(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if thread.Status.WorkspaceID == "" {
		return types.NewErrHttp(http.StatusTooEarly, fmt.Sprintf("no workspace found for assistant %s", id))
	}

	_, err := uploadFileToWorkspace(req.Context(), req, a.gptScript, thread.Status.WorkspaceID, "files/")
	return err
}

func (a *AssistantHandler) DeleteFile(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if thread.Status.WorkspaceID == "" {
		return nil
	}

	return deleteFileFromWorkspaceID(req.Context(), req, a.gptScript, thread.Status.WorkspaceID, "files/")
}

func (a *AssistantHandler) Knowledge(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if len(thread.Status.KnowledgeSetNames) == 0 {
		return req.Write(types.KnowledgeFileList{Items: []types.KnowledgeFile{}})
	}

	return listKnowledgeFiles(req, "", thread.Name, thread.Status.KnowledgeSetNames[0], nil)
}

func (a *AssistantHandler) UploadKnowledge(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if len(thread.Status.KnowledgeSetNames) == 0 {
		return types.NewErrHttp(http.StatusTooEarly, "knowledge set is not available yet")
	}

	ws, err := getWorkspaceFromKnowledgeSet(req, thread.Status.KnowledgeSetNames[0])
	if err != nil {
		return err
	}

	return uploadKnowledgeToWorkspace(req, a.gptScript, ws, "", thread.Name, thread.Status.KnowledgeSetNames[0])
}

func (a *AssistantHandler) DeleteKnowledge(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if len(thread.Status.KnowledgeSetNames) == 0 {
		return types.NewErrHttp(http.StatusTooEarly, fmt.Sprintf("knowledge set is not created yet"))
	}

	return deleteKnowledge(req, req.PathValue("file"), thread.Status.KnowledgeSetNames[0])
}

func appendTools(result *types.AssistantToolList, added map[string]bool, toolsByName map[string]v1.ToolShortDescription, enabled, builtin bool, toolNames []string) {
	for _, toolName := range toolNames {
		if _, ok := added[toolName]; ok {
			continue
		}

		tool, ok := toolsByName[toolName]
		if !ok {
			continue
		}

		newTool := types.AssistantTool{
			ID:          toolName,
			Name:        tool.Name,
			Description: tool.Description,
			Icon:        tool.Metadata["icon"],
			Enabled:     enabled,
			Builtin:     builtin,
		}

		added[toolName] = true
		result.Items = append(result.Items, newTool)
	}
}

func (a *AssistantHandler) AddTool(req api.Context) error {
	var (
		id   = req.PathValue("id")
		tool = req.PathValue("tool")
	)

	agent, err := a.getAgent(req, id)
	if err != nil {
		return err
	}

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	if slices.Contains(thread.Spec.Manifest.Tools, tool) {
		return a.Tools(req)
	}

	if !slices.Contains(agent.Spec.Manifest.AvailableThreadTools, tool) &&
		!slices.Contains(agent.Spec.Manifest.DefaultThreadTools, tool) {
		return types.NewErrBadRequest("tool %s is not available", tool)
	}

	thread.Spec.Manifest.Tools = append(thread.Spec.Manifest.Tools, tool)
	if err := req.Update(&thread); err != nil {
		return err
	}

	return a.Tools(req)
}

func (a *AssistantHandler) RemoveTool(req api.Context) error {
	var (
		id   = req.PathValue("id")
		tool = req.PathValue("tool")
	)

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); err != nil {
		return err
	}

	removed := slices.DeleteFunc(thread.Spec.Manifest.Tools, func(s string) bool {
		return s == tool
	})
	if len(removed) == len(thread.Spec.Manifest.Tools) {
		return types.NewErrNotFound("tool %s not found", tool)
	}
	if err := req.Update(&thread); err != nil {
		return err
	}

	return a.Tools(req)
}

func (a *AssistantHandler) Tools(req api.Context) error {
	var (
		id = req.PathValue("id")
	)

	agent, err := a.getAgent(req, id)
	if err != nil {
		return err
	}

	var thread v1.Thread
	if err := req.Get(&thread, getUserThreadID(id, req.User)); kclient.IgnoreNotFound(err) != nil {
		return err
	}

	enabledTool := make(map[string]bool)
	for _, tool := range thread.Spec.Manifest.Tools {
		enabledTool[tool] = true
	}

	var tools v1.ToolReferenceList
	if err := req.List(&tools); err != nil {
		return err
	}

	toolsByName := make(map[string]v1.ToolShortDescription)
	for _, tool := range tools.Items {
		if tool.Status.Tool != nil {
			toolsByName[tool.Name] = *tool.Status.Tool
		}
	}

	var (
		added  = map[string]bool{}
		result = types.AssistantToolList{
			Items: []types.AssistantTool{},
		}
	)

	appendTools(&result, added, toolsByName, true, true, agent.Spec.Manifest.Tools)

	if thread.Name == "" {
		result.ReadOnly = true
		appendTools(&result, added, toolsByName, true, false, agent.Spec.Manifest.DefaultThreadTools)
	} else {
		appendTools(&result, added, toolsByName, true, false, thread.Spec.Manifest.Tools)
		appendTools(&result, added, toolsByName, false, false, agent.Spec.Manifest.DefaultThreadTools)
	}

	appendTools(&result, added, toolsByName, false, false, agent.Spec.Manifest.AvailableThreadTools)

	sort.Slice(result.Items, func(i, j int) bool {
		return result.Items[i].Name < result.Items[j].Name
	})
	return req.Write(result)
}
