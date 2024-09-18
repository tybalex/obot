package invoke

import (
	"context"

	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/system"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invoker) SystemAction(ctx context.Context, generateName, namespace, tool, input string, env ...string) (*Response, error) {
	thread := v1.Thread{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: system.ThreadPrefix + generateName,
			Namespace:    namespace,
			Finalizers:   []string{v1.ThreadFinalizer},
		},
		Spec: v1.ThreadSpec{
			Input: input,
		},
	}

	if err := i.storage.Create(ctx, &thread); err != nil {
		return nil, err
	}

	return i.SystemActionWithThread(ctx, &thread, tool, input, env...)
}

func (i *Invoker) SystemActionWithThread(ctx context.Context, thread *v1.Thread, tool, input string, env ...string) (*Response, error) {
	return i.createRunFromRemoteTool(ctx, thread, tool, input, runOptions{
		Env: env,
	})
}
