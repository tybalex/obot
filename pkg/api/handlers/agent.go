package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	v2 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AgentHandler struct {
	WorkspaceClient   *client.Client
	WorkspaceProvider string
}

func (a *AgentHandler) Update(ctx context.Context, req api.Request) error {
	var (
		id    = req.Request.PathValue("id")
		agent v2.Agent
	)

	if err := req.Get(&agent, id); err != nil {
		return err
	}

	spec, err := a.parseAgentSpec(ctx, req)
	if err != nil {
		return err
	}

	if spec.Manifest.ID != "" && spec.Manifest.ID != agent.Name {
		return api.NewErrBadRequest("agent ID and ID in manifest do not match %s != %s", agent.Name, spec.Manifest.ID)
	}

	agent.Spec = *spec
	if err := req.Update(&agent); err != nil {
		return err
	}

	return req.JSON(convertAgent(agent, api.GetURLPrefix(req)))
}

func (a *AgentHandler) Delete(ctx context.Context, req api.Request) error {
	var (
		id    = req.Request.PathValue("id")
		agent v2.Agent
	)

	var httpErr *api.ErrHTTP
	if err := req.Get(&agent, id); errors.As(err, &httpErr) && httpErr.Code == http.StatusNotFound {
		return nil
	} else if err != nil {
		return err
	}

	if err := req.Delete(&agent); err != nil {
		return err
	}

	return nil
}

func (a *AgentHandler) parseAgentSpec(ctx context.Context, req api.Request) (*v2.AgentSpec, error) {
	data, err := req.Body()
	if err != nil {
		return nil, err
	}

	var manifest v2.AgentManifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return nil, api.NewErrBadRequest("invalid definition: %v", err)
	}

	return &v2.AgentSpec{
		Manifest:       manifest,
		ManifestSource: string(data),
		Format:         v2.TOMLFormat,
	}, nil
}

func (a *AgentHandler) Create(ctx context.Context, req api.Request) error {
	replace := req.Request.URL.Query().Get("replace") == "true"

	spec, err := a.parseAgentSpec(ctx, req)
	if err != nil {
		return err
	}

	if replace && spec.Manifest.ID == "" {
		return api.NewErrBadRequest("replace requires \"id\" in the manifest to be set")
	}

	var (
		agent     v2.Agent
		httpError *api.ErrHTTP
	)
	if err = req.Get(&agent, spec.Manifest.ID); err != nil {
		if !errors.As(err, &httpError) || httpError.Code != http.StatusNotFound {
			return err
		}
	} else if !replace {
		return apierrors.NewAlreadyExists(v2.SchemeGroupVersion.WithResource("agents").GroupResource(), spec.Manifest.ID)
	}

	agent = v2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Manifest.ID,
			Namespace: req.Namespace(),
		},
		Spec: *spec,
	}

	if agent.Name == "" {
		agent.GenerateName = "a"
	}

	if err = req.Create(&agent); replace && apierrors.IsAlreadyExists(err) {
		err = req.Update(&agent)
		req.ResponseWriter.Header().Set("X-Otto-Replaced", "true")
	}
	if err != nil {
		return err
	}

	req.WriteHeader(http.StatusCreated)
	return req.JSON(convertAgent(agent, api.GetURLPrefix(req)))
}

func convertAgent(agent v2.Agent, prefix string) types.Agent {
	return types.Agent{
		ID:      agent.Name,
		Created: agent.CreationTimestamp.Time,
		Links: map[string]string{
			"invoke": prefix + "/invoke/" + agent.Name,
		},
		Name:        agent.Spec.Manifest.Name,
		Description: agent.Spec.Manifest.Description,
		Manifest:    agent.Spec.Manifest,
	}
}

func (a *AgentHandler) List(_ context.Context, req api.Request) error {
	var agentList v2.AgentList
	if err := req.List(&agentList); err != nil {
		return err
	}

	var resp types.AgentList
	for _, agent := range agentList.Items {
		resp.Items = append(resp.Items, convertAgent(agent, api.GetURLPrefix(req)))
	}

	return req.JSON(resp)
}

func (a *AgentHandler) Files(ctx context.Context, req api.Request) error {
	var (
		id    = req.Request.PathValue("id")
		agent v2.Agent
	)
	if err := req.Get(&agent, id); err != nil {
		return fmt.Errorf("failed to get agent with id %s: %w", id, err)
	}

	return listFiles(ctx, req, a.WorkspaceClient, agent.Status.WorkspaceID)
}

func (a *AgentHandler) UploadFile(ctx context.Context, req api.Request) error {
	var (
		id    = req.Request.PathValue("id")
		agent v2.Agent
	)
	if err := req.Get(&agent, id); err != nil {
		return fmt.Errorf("failed to get agent with id %s: %w", id, err)
	}

	return uploadFile(ctx, req, a.WorkspaceClient, agent.Status.WorkspaceID)
}

func (a *AgentHandler) DeleteFile(ctx context.Context, req api.Request) error {
	var (
		id       = req.Request.PathValue("id")
		filename = req.Request.PathValue("file")
		agent    v2.Agent
	)

	if err := req.Get(&agent, id); err != nil {
		return fmt.Errorf("failed to get agent with id %s: %w", id, err)
	}

	return deleteFile(ctx, req, a.WorkspaceClient, agent.Status.WorkspaceID, filename)
}

func (a *AgentHandler) Knowledge(ctx context.Context, req api.Request) error {
	var (
		id    = req.Request.PathValue("id")
		agent v2.Agent
	)
	if err := req.Get(&agent, id); err != nil {
		return fmt.Errorf("failed to get agent with id %s: %w", id, err)
	}

	return listFiles(ctx, req, a.WorkspaceClient, agent.Status.KnowledgeWorkspaceID)
}

func (a *AgentHandler) UploadKnowledge(ctx context.Context, req api.Request) error {
	var (
		id    = req.Request.PathValue("id")
		agent v2.Agent
	)
	if err := req.Get(&agent, id); err != nil {
		return fmt.Errorf("failed to get agent with id %s: %w", id, err)
	}

	if err := uploadFile(ctx, req, a.WorkspaceClient, agent.Status.KnowledgeWorkspaceID); err != nil {
		return err
	}

	agent.Status.IngestKnowledge = true
	agent.Status.HasKnowledge = true
	return req.Storage.Status().Update(ctx, &agent)
}

func (a *AgentHandler) DeleteKnowledge(ctx context.Context, req api.Request) error {
	var (
		id       = req.Request.PathValue("id")
		filename = req.Request.PathValue("file")
		agent    v2.Agent
	)

	if err := req.Get(&agent, id); err != nil {
		return fmt.Errorf("failed to get agent with id %s: %w", id, err)
	}

	if err := deleteFile(ctx, req, a.WorkspaceClient, agent.Status.KnowledgeWorkspaceID, filename); err != nil {
		return err
	}

	files, err := a.WorkspaceClient.Ls(ctx, agent.Status.KnowledgeWorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace %s: %w", agent.Status.KnowledgeWorkspaceID, err)
	}

	agent.Status.IngestKnowledge = true
	agent.Status.HasKnowledge = len(files) > 0
	return req.Storage.Status().Update(ctx, &agent)
}

func (a *AgentHandler) IngestKnowledge(ctx context.Context, req api.Request) error {
	var (
		id    = req.Request.PathValue("id")
		agent v2.Agent
	)
	if err := req.Get(&agent, id); err != nil {
		return fmt.Errorf("failed to get agent with id %s: %w", id, err)
	}

	files, err := a.WorkspaceClient.Ls(ctx, agent.Status.KnowledgeWorkspaceID)
	if err != nil {
		return err
	}

	req.WriteHeader(http.StatusNoContent)

	if len(files) == 0 && !agent.Status.HasKnowledge {
		return nil
	}

	agent.Status.IngestKnowledge = true
	return req.Storage.Status().Update(ctx, &agent)
}
