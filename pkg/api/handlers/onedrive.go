package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createOneDriveLinks(req api.Context, parentName string, parentObj client.Object) error {
	if err := req.Get(parentObj, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", parentName, err)
	}

	var links []string
	if err := json.NewDecoder(req.Request.Body).Decode(&links); err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}

	oneDriveLinks := &v1.OneDriveLinks{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ol1",
			Namespace:    req.Namespace(),
		},
		Spec: v1.OnedriveLinksSpec{
			SharedLinks: links,
		},
	}

	switch parentObj.(type) {
	case *v1.Agent:
		oneDriveLinks.Spec.AgentName = parentName
	case *v1.Workflow:
		oneDriveLinks.Spec.WorkflowName = parentName
	default:
		return fmt.Errorf("unknown parent object type: %T", parentObj)
	}

	if err := req.Create(oneDriveLinks); err != nil {
		return fmt.Errorf("failed to create OneDrive links: %w", err)
	}

	req.WriteHeader(http.StatusCreated)

	return nil
}

func updateOneDriveLinks(req api.Context, linksID, parentName string, parentObj client.Object) error {
	if err := req.Get(parentObj, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", parentName, err)
	}

	var oneDriveLinks v1.OneDriveLinks
	if err := req.Get(&oneDriveLinks, linksID); err != nil {
		return fmt.Errorf("failed to get OneDrive links with id %s: %w", linksID, err)
	}

	if oneDriveLinks.Spec.AgentName != parentName && oneDriveLinks.Spec.WorkflowName != parentName {
		return fmt.Errorf("OneDrive links agent name %q does not match provided agent name %q", oneDriveLinks.Spec.AgentName, parentName)
	}

	var links []string
	if err := json.NewDecoder(req.Request.Body).Decode(&links); err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}

	oneDriveLinks.Spec.SharedLinks = links

	if err := req.Update(&oneDriveLinks); err != nil {
		return fmt.Errorf("failed to update OneDrive links: %w", err)
	}

	return nil
}

func reSyncOneDriveLinks(req api.Context, linksID, parentName string, parentObj client.Object) error {
	if err := req.Get(parentObj, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", parentName, err)
	}

	var oneDriveLinks v1.OneDriveLinks
	if err := req.Get(&oneDriveLinks, linksID); err != nil {
		return fmt.Errorf("failed to get OneDrive links with id %s: %w", linksID, err)
	}

	if oneDriveLinks.Spec.AgentName != parentName && oneDriveLinks.Spec.WorkflowName != parentName {
		return fmt.Errorf("OneDrive links agent name %q does not match provided agent name %q", oneDriveLinks.Spec.AgentName, parentName)
	}

	oneDriveLinks.Status.ObservedGeneration = 0
	oneDriveLinks.Status.ThreadName = ""
	oneDriveLinks.Status.RunName = ""
	if err := req.Storage.Status().Update(req.Context(), &oneDriveLinks); err != nil {
		return fmt.Errorf("failed to resync OneDrive links: %w", err)
	}

	return nil
}

func getOneDriveLinksForParent(req api.Context, parentName string, parentObj client.Object) error {
	if err := req.Get(parentObj, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", parentName, err)
	}

	var oneDriveLinks v1.OneDriveLinksList
	if err := req.List(&oneDriveLinks); err != nil {
		return fmt.Errorf("failed to get OneDrive links with id %s: %w", parentName, err)
	}

	var resp v1.OneDriveLinksList
	for _, link := range oneDriveLinks.Items {
		if link.Spec.WorkflowName == parentName || link.Spec.AgentName == parentName {
			resp.Items = append(resp.Items, link)
		}
	}

	return req.Write(&resp)
}

func deleteOneDriveLinks(req api.Context, linksID, parentName string, parentObj client.Object) error {
	if err := req.Get(parentObj, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", parentName, err)
	}

	var (
		httpErr       *api.ErrHTTP
		oneDriveLinks v1.OneDriveLinks
	)
	if err := req.Get(&oneDriveLinks, linksID); errors.As(err, &httpErr) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get OneDrive links with id %s: %w", linksID, err)
	}

	if oneDriveLinks.Spec.AgentName != parentName && oneDriveLinks.Spec.WorkflowName != parentName {
		return fmt.Errorf("OneDrive links name %q does not match provided agent name %q", oneDriveLinks.Spec.AgentName, parentName)
	}

	if err := req.Delete(&oneDriveLinks); err != nil {
		return fmt.Errorf("failed to delete OneDrive links: %w", err)
	}

	return nil
}
