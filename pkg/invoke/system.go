package invoke

import (
	"context"

	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.gptscript.ai/v1"
)

func (i *Invoker) SystemAction(ctx context.Context, generateName, agentName, namespace, tool, input string, env ...string) (*Response, error) {
	thread, err := i.NewThread(ctx, i.uncached, namespace, NewThreadOptions{
		AgentName:          agentName,
		ThreadGenerateName: generateName,
	})
	if err != nil {
		return nil, err
	}

	return i.SystemActionWithThread(ctx, thread, tool, input, env...)
}

func (i *Invoker) SystemActionWithThread(ctx context.Context, thread *v1.Thread, tool, input string, env ...string) (*Response, error) {
	return i.createRunFromRemoteTool(ctx, i.uncached, thread, tool, input, runOptions{
		AgentName: thread.Spec.AgentName,
		Env:       env,
	})
}
