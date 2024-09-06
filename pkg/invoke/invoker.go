package invoke

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/agents"
	"github.com/gptscript-ai/otto/pkg/gz"
	"github.com/gptscript-ai/otto/pkg/jwt"
	"github.com/gptscript-ai/otto/pkg/storage"
	v2 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Invoker struct {
	storage                 storage.Client
	gptClient               *gptscript.GPTScript
	tokenService            *jwt.TokenService
	workspaceClient         *wclient.Client
	threadWorkspaceProvider string
}

func NewInvoker(storage storage.Client, gptClient *gptscript.GPTScript, tokenService *jwt.TokenService, workspaceClient *wclient.Client) *Invoker {
	return &Invoker{
		storage:                 storage,
		gptClient:               gptClient,
		tokenService:            tokenService,
		workspaceClient:         workspaceClient,
		threadWorkspaceProvider: "directory",
	}
}

type Response struct {
	Run    *v2.Run
	Thread *v2.Thread
	Events <-chan v2.Progress
}

type Options struct {
	ThreadName string
}

func getWorkspace(thread *v2.Thread) string {
	_, path, _ := strings.Cut(thread.Spec.WorkspaceID, "://")
	return path
}

func (i *Invoker) getThread(ctx context.Context, agent *v2.Agent, input, threadName string) (*v2.Thread, error) {
	var (
		thread     v2.Thread
		createName string
	)
	if threadName != "" {
		err := i.storage.Get(ctx, router.Key(agent.Namespace, threadName), &thread)
		if apierror.IsNotFound(err) {
			createName = threadName
		} else {
			return &thread, err
		}
	}

	workspaceID, err := i.workspaceClient.Create(ctx, i.threadWorkspaceProvider, agent.Spec.WorkspaceID)
	if err != nil {
		return nil, err
	}

	thread = v2.Thread{
		ObjectMeta: metav1.ObjectMeta{
			Name:         createName,
			GenerateName: "t",
			Namespace:    agent.Namespace,
		},
		Spec: v2.ThreadSpec{
			AgentName:   agent.Name,
			Input:       input,
			WorkspaceID: workspaceID,
		},
	}
	if err := i.storage.Create(ctx, &thread); err != nil {
		// If creating the thread fails, then ensure that the workspace is cleaned up, too.
		return nil, errors.Join(err, i.workspaceClient.Rm(ctx, thread.Spec.WorkspaceID))
	}
	return &thread, nil
}

func (i *Invoker) getChatState(ctx context.Context, run *v2.Run) (result string, _ error) {
	if run.Spec.PreviousRunName == "" {
		return "", nil
	}

	var lastRun v2.RunState
	if err := i.storage.Get(ctx, router.Key(run.Namespace, run.Spec.PreviousRunName), &lastRun); apierror.IsNotFound(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	err := gz.Decompress(&result, lastRun.Spec.ChatState)
	return result, err
}

func (i *Invoker) Invoke(ctx context.Context, agent *v2.Agent, input string, opts ...Options) (*Response, error) {
	var opt Options
	for _, o := range opts {
		if o.ThreadName != "" {
			opt.ThreadName = o.ThreadName
		}
	}

	thread, err := i.getThread(ctx, agent, input, opt.ThreadName)
	if err != nil {
		return nil, err
	}

	tools, err := agents.Render(ctx, i.storage, agent.Namespace, agent.Spec.Manifest)
	if err != nil {
		return nil, err
	}

	if len(agent.Spec.Manifest.Params) == 0 {
		data := map[string]any{}
		if err := json.Unmarshal([]byte(input), &data); err == nil {
			if msg, ok := data[agents.DefaultAgentParams[0]].(string); ok && len(data) == 1 && msg != "" {
				input = msg
			}
		}
	}

	var run = v2.Run{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "r",
			Namespace:    thread.Namespace,
		},
		Spec: v2.RunSpec{
			ThreadName:      thread.Name,
			AgentName:       agent.Name,
			PreviousRunName: thread.Status.LastRunName,
			Input:           input,
		},
	}

	if err := i.storage.Create(ctx, &run); err != nil {
		return nil, err
	}

	chatState, err := i.getChatState(ctx, &run)
	if err != nil {
		return nil, err
	}

	token, err := i.tokenService.NewToken(jwt.TokenContext{
		RunID:    run.Name,
		ThreadID: thread.Name,
		AgentID:  agent.Name,
		Scope:    agent.Namespace,
	})

	runResp, err := i.gptClient.Evaluate(ctx, gptscript.Options{
		GlobalOptions: gptscript.GlobalOptions{
			Env: append(os.Environ(),
				"OTTO_TOKEN="+token,
				"OTTO_RUN_ID="+run.Name,
				"OTTO_THREAD_ID="+thread.Name,
				"OTTO_AGENT_ID="+agent.Name),
		},
		Input:         input,
		Workspace:     getWorkspace(thread),
		ChatState:     chatState,
		IncludeEvents: true,
	}, tools...)
	if err != nil {
		return nil, err
	}

	var events = make(chan v2.Progress)

	go i.stream(ctx, events, thread, &run, runResp)

	return &Response{
		Run:    &run,
		Thread: thread,
		Events: events,
	}, nil
}

func (i *Invoker) saveState(ctx context.Context, thread *v2.Thread, run *v2.Run, runResp *gptscript.Run, retErr error) error {
	var (
		runStateSpec v2.RunStateSpec
		runChanged   bool
		err          error
	)

	runStateSpec.ThreadName = run.Spec.ThreadName

	if prg := runResp.Program(); prg != nil {
		runStateSpec.Program, err = gz.Compress(prg)
		if err != nil {
			return err
		}
	}

	runStateSpec.CallFrame, err = gz.Compress(runResp.Calls())
	if err != nil {
		return err
	}

	if chatState := runResp.ChatState(); chatState != "" {
		runStateSpec.ChatState, err = gz.Compress(chatState)
		if err != nil {
			return err
		}
	}

	var runState v2.RunState
	if err := i.storage.Get(ctx, router.Key(run.Namespace, run.Name), &runState); apierror.IsNotFound(err) {
		runState = v2.RunState{
			ObjectMeta: metav1.ObjectMeta{
				Name:      run.Name,
				Namespace: run.Namespace,
			},
			Spec: runStateSpec,
		}
		if err := i.storage.Create(ctx, &runState); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		if !bytes.Equal(runState.Spec.CallFrame, runStateSpec.CallFrame) ||
			!bytes.Equal(runState.Spec.ChatState, runStateSpec.ChatState) {
			if err := i.storage.Update(ctx, &runState); err != nil {
				return err
			}
		}
	}

	state := runResp.State()

	if run.Status.State != state {
		run.Status.State = state
		runChanged = true
	}

	var final bool
	switch state {
	case gptscript.Error:
		final = true
		errString := runResp.ErrorOutput()
		if errString == "" {
			errString = runResp.Err().Error()
		}
		if run.Status.Error != errString {
			run.Status.Error = errString
			runChanged = true
		}
	case gptscript.Continue, gptscript.Finished:
		final = true
		text, err := runResp.Text()
		if err != nil {
			// this should never happen because gptscript.Error would have been set
			panic(err)
		}
		if run.Status.Output != text {
			run.Status.Output = text
			runChanged = true
		}
	}

	if runChanged {
		if err := i.storage.Status().Update(ctx, run); err != nil {
			return err
		}
	}

	if final && thread.Status.LastRunName != run.Name {
		thread.Status.LastRunName = run.Name
		thread.Status.LastRunState = run.Status.State
		thread.Status.LastRunOutput = run.Status.Output
		thread.Status.LastRunError = run.Status.Error
		if err := i.storage.Status().Update(ctx, thread); err != nil {
			return err
		}
	}

	return retErr
}

func (i *Invoker) stream(ctx context.Context, events chan v2.Progress, thread *v2.Thread, run *v2.Run, runResp *gptscript.Run) (retErr error) {
	defer close(events)

	var (
		runEvent = runResp.Events()
		wg       sync.WaitGroup
		prg      gptscript.Program
	)
	defer func() {
		retErr = i.saveState(ctx, thread, run, runResp, retErr)
	}()

	defer wg.Wait()

	saveCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-saveCtx.Done():
				return
			case <-time.After(2 * time.Second):
				_ = i.saveState(ctx, thread, run, runResp, nil)
			}
		}
	}()

	var (
		lastPrint = map[string][]gptscript.Output{}
	)

	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	w, err := i.storage.Watch(watchCtx, &v2.RunStateList{}, &client.ListOptions{
		Namespace: run.Namespace,
	})
	if err != nil {
		return err
	}

	watchEvents := w.ResultChan()
	defer func() {
		for range watchEvents {
		}
	}()
	defer w.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watchEvents:
			if !ok {
				watchEvents = nil
				continue
			}
			runState := event.Object.(*v2.RunState)
			if strings.HasPrefix(runState.Spec.ThreadName, thread.Name+".") {
				//printSubCall(runState, lastPrint, events)
			}
		case frame, ok := <-runEvent:
			if !ok {
				return runResp.Err()
			}

			if frame.Run != nil {
				if len(frame.Run.Program.ToolSet) > 0 {
					prg = frame.Run.Program
				}
			} else if frame.Call != nil {
				switch frame.Call.Type {
				case gptscript.EventTypeCallProgress, gptscript.EventTypeCallFinish:
					if frame.Call.ToolCategory == gptscript.NoCategory && frame.Call.Tool.Chat {
						printCall(prg, frame.Call, lastPrint, events)
					}
				}
			}
		}
	}
}

func printString(out chan v2.Progress, last, current string) {
	current = strings.TrimPrefix(current, "Waiting for model response...")
	current, _, _ = strings.Cut(current, "<tool call> ")
	if strings.HasPrefix(current, last) {
		out <- v2.Progress{
			Content: current[len(last):],
		}
	} else {
		out <- v2.Progress{
			Content: current,
		}
	}
}

func printSubCall(runState *v2.RunState, lastPrint map[string][]gptscript.Output, out chan v2.Progress) {
	var (
		prg   gptscript.Program
		calls = map[string]gptscript.CallFrame{}
	)
	if err := gz.Decompress(&calls, runState.Spec.CallFrame); err != nil {
		return
	}
	if err := gz.Decompress(&prg, runState.Spec.Program); err != nil {
		return
	}
	for _, call := range calls {
		if call.ParentID == "" {
			printCall(prg, &call, lastPrint, out)
		}
	}
}

func printCall(prg gptscript.Program, call *gptscript.CallFrame, lastPrint map[string][]gptscript.Output, out chan v2.Progress) {
	lastOutputs := lastPrint[call.ID]
	for i, currentOutput := range call.Output {
		for i >= len(lastOutputs) {
			lastOutputs = append(lastOutputs, gptscript.Output{})
		}
		last := lastOutputs[i]

		if last.Content != currentOutput.Content {
			printString(out, last.Content, currentOutput.Content)
			last.Content = currentOutput.Content
		}

		if last.SubCalls == nil {
			last.SubCalls = map[string]gptscript.Call{}
		}

		for subCallID, subCall := range currentOutput.SubCalls {
			lastSubCall, ok := last.SubCalls[subCallID]
			if !ok || lastSubCall != subCall {
				tool, ok := prg.ToolSet[subCall.ToolID]
				if ok {
					out <- v2.Progress{
						Tool: v2.ToolProgress{
							Name:        tool.Name,
							Description: tool.Description,
							Input:       subCall.Input,
						},
					}
				}
			}
			last.SubCalls[subCallID] = subCall
		}

		lastOutputs[i] = currentOutput
	}

	lastPrint[call.ID] = lastOutputs
}
