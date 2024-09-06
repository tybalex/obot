package handlers

import (
	"context"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	v2 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AgentHandler struct {
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

	if err := req.Get(&agent, id); apierrors.IsNotFound(err) {
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

	var manifest v2.Manifest
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

	agent := v2.Agent{
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

	req.ResponseWriter.WriteHeader(http.StatusCreated)
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
