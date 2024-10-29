package knowledgefile

import (
	"context"
	"fmt"
	"path"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/acorn-io/baaah/pkg/typed"
	"github.com/otto8-ai/otto8/apiclient/types"
	"github.com/otto8-ai/otto8/pkg/invoke"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/otto8-ai/otto8/pkg/system"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	invoker *invoke.Invoker
}

func New(invoker *invoke.Invoker) *Handler {
	return &Handler{
		invoker: invoker,
	}
}

func shouldReIngest(file *v1.KnowledgeFile) bool {
	return file.Spec.IngestGeneration > file.Status.IngestGeneration ||
		file.Spec.UpdatedAt != file.Status.UpdatedAt ||
		file.Spec.Checksum != file.Status.Checksum ||
		file.Spec.URL != file.Status.URL
}

func outputFile(filename string) string {
	return path.Join(".conversion", filename+".txt")
}

func getThread(ctx context.Context, c kclient.Client, ks *v1.KnowledgeSet, source *v1.KnowledgeSource) (*v1.Thread, error) {
	var thread v1.Thread
	if source.Status.ThreadName != "" {
		return &thread, c.Get(ctx, router.Key(ks.Namespace, source.Status.ThreadName), &thread)
	}
	return &thread, c.Get(ctx, router.Key(ks.Namespace, ks.Status.ThreadName), &thread)
}

func (h *Handler) IngestFile(req router.Request, _ router.Response) error {
	file := req.Object.(*v1.KnowledgeFile)

	var source v1.KnowledgeSource
	if file.Spec.KnowledgeSourceName != "" {
		if err := req.Client.Get(req.Ctx, router.Key(file.Namespace, file.Spec.KnowledgeSourceName), &source); err != nil {
			// NotFound is fine, Cleanup handler will handle things and end up deleting this file
			return kclient.IgnoreNotFound(err)
		}
	}

	var ks v1.KnowledgeSet
	if err := req.Client.Get(req.Ctx, router.Key(file.Namespace, file.Spec.KnowledgeSetName), &ks); err != nil {
		// NotFound is fine, Cleanup handler will handle things and end up deleting this file
		return kclient.IgnoreNotFound(err)
	}

	thread, err := getThread(req.Ctx, req.Client, &ks, &source)
	if err != nil {
		return err
	}

	if file.Status.State == "" {
		// We don't need to save it right now, has no real impact since the API turns "" into pending
		file.Status.State = types.KnowledgeFileStatePending
	}

	if file.Status.State == types.KnowledgeFileStateIngesting {
		// Resuming from failed system restart, go back to pending
		file.Status.State = types.KnowledgeFileStatePending
		if err := req.Client.Status().Update(req.Ctx, file); err != nil {
			return err
		}
	}

	if file.Status.State.IsTerminal() && !shouldReIngest(file) {
		return nil
	}

	// We should be pending at this point, if not update to that state
	if file.Status.State != types.KnowledgeFileStatePending {
		file.Status.State = types.KnowledgeFileStatePending
		if err := req.Client.Status().Update(req.Ctx, file); err != nil {
			return err
		}
	}

	// Check approval
	if file.Spec.Approved == nil {
		if source.Spec.Manifest.AutoApprove != nil && *source.Spec.Manifest.AutoApprove {
			file.Spec.Approved = typed.Pointer(true)
			if err := req.Client.Update(req.Ctx, file); err != nil {
				return err
			}
		}
	}

	if file.Spec.Approved == nil || !*file.Spec.Approved {
		// Not approved, wait for user action
		return nil
	}

	if err := h.ingest(req.Ctx, req.Client, file, &ks, thread); err != nil {
		file.Status.State = types.KnowledgeFileStateError
		file.Status.Error = err.Error()
	} else {
		file.Status.State = types.KnowledgeFileStateIngested
		file.Status.Error = ""
	}

	file.Status.LastIngestionEndTime = metav1.Now()
	file.Status.URL = file.Spec.URL
	file.Status.UpdatedAt = file.Spec.UpdatedAt
	file.Status.Checksum = file.Spec.Checksum
	file.Status.IngestGeneration = file.Spec.IngestGeneration
	return req.Client.Status().Update(req.Ctx, file)
}

func (h *Handler) ingest(ctx context.Context, client kclient.Client, file *v1.KnowledgeFile,
	ks *v1.KnowledgeSet, thread *v1.Thread) error {

	task, err := h.invoker.SystemTask(ctx, thread, system.KnowledgeLoadTool, map[string]any{
		"input":  file.Spec.FileName,
		"output": outputFile(file.Spec.FileName),
		"metadata": map[string]string{
			"url": file.Spec.URL,
		},
	})
	if err != nil {
		return err
	}
	defer task.Close()

	file.Status.State = types.KnowledgeFileStateIngesting
	file.Status.Error = ""
	file.Status.LastIngestionStartTime = metav1.Now()
	file.Status.LastIngestionEndTime = metav1.Time{}
	file.Status.ThreadName = task.Thread.Name
	file.Status.RunName = task.Run.Name
	if err := client.Status().Update(ctx, file); err != nil {
		return err
	}

	result, err := task.Result(ctx)
	if err != nil {
		return err
	}
	if result.Error != "" {
		return fmt.Errorf("failed to convert file: %s", result.Error)
	}

	task, err = h.invoker.SystemTask(ctx, thread, system.KnowledgeIngestTool, map[string]any{
		"input":   outputFile(file.Spec.FileName),
		"dataset": ks.Namespace + "/" + ks.Name,
	})
	if err != nil {
		return err
	}
	defer task.Close()

	file.Status.RunName = task.Run.Name
	if err := client.Status().Update(ctx, file); err != nil {
		return err
	}

	result, err = task.Result(ctx)
	if err != nil {
		return err
	}
	if result.Error != "" {
		return fmt.Errorf("failed to ingest file: %s", result.Error)
	}

	return nil
}

func (h *Handler) Cleanup(req router.Request, _ router.Response) error {
	file := req.Object.(*v1.KnowledgeFile)

	var ks v1.KnowledgeSet
	if err := req.Client.Get(req.Ctx, router.Key(file.Namespace, file.Spec.KnowledgeSetName), &ks); apierror.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	var thread v1.Thread
	if err := req.Client.Get(req.Ctx, router.Key(ks.Namespace, file.Status.ThreadName), &thread); apierror.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	task, err := h.invoker.SystemTask(req.Ctx, &thread, system.KnowledgeDeleteFileTool, map[string]any{
		"file":    file.Spec.FileName,
		"dataset": ks.Namespace + "/" + ks.Name,
	})
	if err != nil {
		return err
	}
	defer task.Close()

	_, err = task.Result(req.Ctx)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge file: %w", err)
	}
	return nil
}
