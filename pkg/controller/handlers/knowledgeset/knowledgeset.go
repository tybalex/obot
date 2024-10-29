package knowledgeset

import (
	"context"
	"fmt"
	"strings"

	"github.com/acorn-io/baaah/pkg/name"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/otto8-ai/otto8/pkg/aihelper"
	"github.com/otto8-ai/otto8/pkg/create"
	"github.com/otto8-ai/otto8/pkg/invoke"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/otto8-ai/otto8/pkg/system"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	aiHelper *aihelper.AIHelper
	invoker  *invoke.Invoker
}

func New(aiHelper *aihelper.AIHelper, invoker *invoke.Invoker) *Handler {
	return &Handler{
		aiHelper: aiHelper,
		invoker:  invoker,
	}
}

func (h *Handler) GenerateDataDescription(req router.Request, _ router.Response) error {
	return nil
}

func generatePrompt(files v1.KnowledgeFileList) string {
	var (
		prompt    string
		fileNames = make([]string, 0, len(files.Items))
	)

	for _, file := range files.Items {
		fileNames = append(fileNames, "- "+file.Spec.FileName)
	}

	fileText := strings.Join(fileNames, "\n")
	if len(fileText) > 50000 {
		fileText = fileText[:50000]
	}

	prompt = "The following files are in this knowledge set:\n" + fileText
	prompt += "\n\nGenerate a 50 word description of the data in the knowledge set that would help a" +
		" reader understand why they might want to search this knowledge set. Be precise and concise."
	return prompt
}

func createWorkspace(ctx context.Context, c kclient.Client, ks *v1.KnowledgeSet) error {
	if ks.Status.WorkspaceName != "" {
		return nil
	}

	ws := &v1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name.SafeConcatName(system.WorkspacePrefix, ks.Name),
			Namespace:  ks.Namespace,
			Finalizers: []string{v1.WorkspaceFinalizer},
		},
		Spec: v1.WorkspaceSpec{
			KnowledgeSetName: ks.Name,
		},
	}
	err := create.OrGet(ctx, c, ws)
	if err != nil {
		return err
	}

	ks.Status.WorkspaceName = ws.Name
	return c.Status().Update(ctx, ks)
}

func (h *Handler) createThread(ctx context.Context, c kclient.Client, ks *v1.KnowledgeSet) error {
	thread := &v1.Thread{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.SafeConcatName(system.ThreadPrefix, ks.Name),
			Namespace: ks.Namespace,
		},
		Spec: v1.ThreadSpec{
			KnowledgeSetName: ks.Name,
			WorkspaceName:    ks.Status.WorkspaceName,
			SystemTask:       true,
		},
	}
	// Threads are special because we assume users might delete them randomly
	err := create.IfNotExists(ctx, c, thread)
	if err != nil {
		return err
	}

	if ks.Status.ThreadName == "" {
		ks.Status.ThreadName = thread.Name
		return c.Status().Update(ctx, ks)
	}
	return nil
}

func (h *Handler) CreateWorkspace(req router.Request, _ router.Response) error {
	ks := req.Object.(*v1.KnowledgeSet)

	if err := createWorkspace(req.Ctx, req.Client, ks); err != nil {
		return err
	}

	return h.createThread(req.Ctx, req.Client, ks)
}

func (h *Handler) Cleanup(req router.Request, _ router.Response) error {
	ks := req.Object.(*v1.KnowledgeSet)
	if ks.Status.ThreadName == "" {
		return nil
	}

	var thread v1.Thread
	if err := req.Client.Get(req.Ctx, router.Key(ks.Namespace, ks.Status.ThreadName), &thread); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	task, err := h.invoker.SystemTask(req.Ctx, &thread, system.KnowledgeDeleteTool, map[string]any{
		"dataset": ks.Namespace + "/" + ks.Name,
	})
	if err != nil {
		return err
	}
	defer task.Close()

	_, err = task.Result(req.Ctx)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge set: %w", err)
	}
	return nil
}
