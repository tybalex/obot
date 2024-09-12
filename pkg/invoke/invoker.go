package invoke

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/gz"
	"github.com/gptscript-ai/otto/pkg/jwt"
	"github.com/gptscript-ai/otto/pkg/render"
	"github.com/gptscript-ai/otto/pkg/storage"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/system"
	"github.com/gptscript-ai/otto/pkg/workspace"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Invoker struct {
	storage                 storage.Client
	gptClient               *gptscript.GPTScript
	tokenService            *jwt.TokenService
	workspaceClient         *wclient.Client
	threadWorkspaceProvider string
	knowledgeTool           string
}

func NewInvoker(storage storage.Client, gptClient *gptscript.GPTScript, tokenService *jwt.TokenService, workspaceClient *wclient.Client, knowledgeTool string) *Invoker {
	return &Invoker{
		storage:                 storage,
		gptClient:               gptClient,
		tokenService:            tokenService,
		workspaceClient:         workspaceClient,
		threadWorkspaceProvider: "directory",
		knowledgeTool:           knowledgeTool,
	}
}

type Response struct {
	Run    *v1.Run
	Thread *v1.Thread
	Events <-chan v1.Progress
}

func (r *Response) Wait() {
	for range r.Events {
	}
}

type Options struct {
	ThreadName string
}

func (i *Invoker) getThread(ctx context.Context, agent *v1.Agent, input, threadName string) (*v1.Thread, error) {
	var (
		thread     v1.Thread
		createName string
	)
	if threadName != "" {
		err := i.storage.Get(ctx, router.Key(agent.Namespace, threadName), &thread)
		if apierror.IsNotFound(err) {
			if system.IsThreadID(threadName) {
				createName = threadName
			} else {
				return nil, err
			}
		} else {
			return &thread, err
		}
	}

	workspaceID, err := i.workspaceClient.Create(ctx, i.threadWorkspaceProvider, agent.Status.WorkspaceID)
	if err != nil {
		return nil, err
	}

	knowledgeWorkspaceID, err := i.workspaceClient.Create(ctx, i.threadWorkspaceProvider, agent.Status.KnowledgeWorkspaceID)
	if err != nil {
		return nil, err
	}

	thread = v1.Thread{
		ObjectMeta: metav1.ObjectMeta{
			Name:         createName,
			GenerateName: "t1",
			Namespace:    agent.Namespace,
			Finalizers:   []string{v1.ThreadFinalizer},
		},
		Spec: v1.ThreadSpec{
			AgentName:            agent.Name,
			Input:                input,
			WorkspaceID:          workspaceID,
			KnowledgeWorkspaceID: knowledgeWorkspaceID,
		},
	}
	if err := i.storage.Create(ctx, &thread); err != nil {
		// If creating the thread fails, then ensure that the workspace is cleaned up, too.
		return nil, errors.Join(err, i.workspaceClient.Rm(ctx, thread.Spec.WorkspaceID))
	}
	return &thread, nil
}

func (i *Invoker) getChatState(ctx context.Context, run *v1.Run) (result string, _ error) {
	if run.Spec.PreviousRunName == "" {
		return "", nil
	}

	for {
		// look for the last valid state
		var previousRun v1.Run
		if err := i.storage.Get(ctx, router.Key(run.Namespace, run.Spec.PreviousRunName), &previousRun); err != nil {
			return "", err
		}
		if previousRun.Status.State == gptscript.Continue {
			break
		}
		if previousRun.Spec.PreviousRunName == "" {
			return "", nil
		}
		run = &previousRun
	}

	var lastRun v1.RunState
	if err := i.storage.Get(ctx, router.Key(run.Namespace, run.Spec.PreviousRunName), &lastRun); apierror.IsNotFound(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	err := gz.Decompress(&result, lastRun.Spec.ChatState)
	return result, err
}

func (i *Invoker) Agent(ctx context.Context, agent *v1.Agent, input string, opts ...Options) (*Response, error) {
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

	tools, extraEnv, err := render.Agent(ctx, i.storage, agent, render.AgentOptions{
		Thread:        thread,
		KnowledgeTool: i.knowledgeTool,
	})
	if err != nil {
		return nil, err
	}

	if len(agent.Spec.Manifest.Params) == 0 {
		data := map[string]any{}
		if err := json.Unmarshal([]byte(input), &data); err == nil {
			if msg, ok := data[render.DefaultAgentParams[0]].(string); ok && len(data) == 1 && msg != "" {
				input = msg
			}
		}
	}

	return i.createRunFromTools(ctx, thread, tools, input, runOptions{
		AgentName: agent.Name,
		Env:       extraEnv,
	})
}

type runOptions struct {
	AgentName        string
	WorkflowName     string
	WorkflowStepName string
	Env              []string
}

func (i *Invoker) createRunFromTools(ctx context.Context, thread *v1.Thread, tools []gptscript.ToolDef, input string, opts runOptions) (*Response, error) {
	return i.createRun(ctx, thread, input, opts, tools)
}

func (i *Invoker) createRunFromRemoteTool(ctx context.Context, thread *v1.Thread, tool, input string, opts runOptions) (*Response, error) {
	return i.createRun(ctx, thread, input, opts, tool)
}

// createRun is a low-level method that creates a Run object from a list of tools or a remote tool.
// Callers should use createRunFromTools or createRunFromRemoteTool instead.
func (i *Invoker) createRun(ctx context.Context, thread *v1.Thread, input string, opts runOptions, tool any) (*Response, error) {
	var run = v1.Run{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "r1",
			Namespace:    thread.Namespace,
			Finalizers:   []string{v1.RunFinalizer},
		},
		Spec: v1.RunSpec{
			ThreadName:       thread.Name,
			AgentName:        opts.AgentName,
			WorkflowStepName: opts.WorkflowStepName,
			PreviousRunName:  thread.Status.LastRunName,
			Input:            input,
			Env:              opts.Env,
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
		RunID:          run.Name,
		ThreadID:       thread.Name,
		AgentID:        opts.AgentName,
		WorkflowID:     opts.WorkflowName,
		WorkflowStepID: opts.WorkflowStepName,
		Scope:          thread.Namespace,
	})

	options := gptscript.Options{
		GlobalOptions: gptscript.GlobalOptions{
			Env: append(opts.Env,
				"OTTO_TOKEN="+token,
				"OTTO_RUN_ID="+run.Name,
				"OTTO_THREAD_ID="+thread.Name,
				"OTTO_WORKFLOW_ID="+opts.WorkflowName,
				"OTTO_WORKFLOW_STEP_ID="+opts.WorkflowStepName,
				"OTTO_AGENT_ID="+opts.AgentName,
			),
		},
		Input:           input,
		Workspace:       workspace.GetDir(thread.Spec.WorkspaceID),
		ChatState:       chatState,
		IncludeEvents:   true,
		ForceSequential: true,
	}

	var runResp *gptscript.Run
	switch t := tool.(type) {
	case gptscript.ToolDef:
		runResp, err = i.gptClient.Evaluate(ctx, options, t)
	case []gptscript.ToolDef:
		runResp, err = i.gptClient.Evaluate(ctx, options, t...)
	case string:
		runResp, err = i.gptClient.Run(ctx, t, options)
	default:
		return nil, fmt.Errorf("invalid tool type: %T", tool)
	}
	if err != nil {
		return nil, err
	}

	var events = make(chan v1.Progress)
	go i.stream(ctx, events, thread, &run, runResp)

	return &Response{
		Run:    &run,
		Thread: thread,
		Events: events,
	}, nil
}

func (i *Invoker) saveState(ctx context.Context, thread *v1.Thread, run *v1.Run, runResp *gptscript.Run, retErr error) error {
	for j := 0; j < 3; j++ {
		err := i.doSaveState(ctx, thread, run, runResp, retErr)
		if err == nil {
			return retErr
		}
		if !apierror.IsConflict(err) {
			return errors.Join(err, retErr)
		}
		// reload
		if err := i.storage.Get(ctx, router.Key(run.Namespace, run.Name), run); err != nil {
			return errors.Join(err, retErr)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("failed to save state after 3 retries: %w", retErr)
}

func (i *Invoker) doSaveState(ctx context.Context, thread *v1.Thread, run *v1.Run, runResp *gptscript.Run, retErr error) error {
	var (
		runStateSpec v1.RunStateSpec
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

	var runState v1.RunState
	if err := i.storage.Get(ctx, router.Key(run.Namespace, run.Name), &runState); apierror.IsNotFound(err) {
		runState = v1.RunState{
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

func (i *Invoker) stream(ctx context.Context, events chan v1.Progress, thread *v1.Thread, run *v1.Run, runResp *gptscript.Run) (retErr error) {
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

	for {
		select {
		case <-ctx.Done():
			return nil
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

func printString(out chan v1.Progress, last, current string) {
	current = strings.TrimPrefix(current, "Waiting for model response...")
	current, _, _ = strings.Cut(current, "<tool call> ")
	if strings.HasPrefix(current, last) {
		out <- v1.Progress{
			Content: current[len(last):],
		}
	} else {
		out <- v1.Progress{
			Content: current,
		}
	}
}

func printSubCall(runState *v1.RunState, lastPrint map[string][]gptscript.Output, out chan v1.Progress) {
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

func printCall(prg gptscript.Program, call *gptscript.CallFrame, lastPrint map[string][]gptscript.Output, out chan v1.Progress) {
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
					out <- v1.Progress{
						Tool: v1.ToolProgress{
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
