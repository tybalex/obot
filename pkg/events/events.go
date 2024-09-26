package events

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/acorn-io/baaah/pkg/typed"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/apiclient/types"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/workflow"
	"github.com/gptscript-ai/otto/pkg/gz"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Emitter struct {
	client        kclient.WithWatch
	liveStates    map[string][]liveState
	liveStateLock sync.RWMutex
	liveBroadcast *sync.Cond
}

func NewEmitter(client kclient.WithWatch) *Emitter {
	e := &Emitter{
		client:     client,
		liveStates: map[string][]liveState{},
	}
	e.liveBroadcast = sync.NewCond(&e.liveStateLock)
	return e
}

type liveState struct {
	Prg      *gptscript.Program
	Frames   Frames
	Progress *types.Progress
	Done     bool
}

type WatchOptions struct {
	History     bool
	LastRunName string
	ThreadName  string
	Follow      bool
	Run         *v1.Run
}

type Frames map[string]gptscript.CallFrame

type callFramePrintState struct {
	Outputs      []gptscript.Output
	InputPrinted bool
}

type printState struct {
	frames          map[string]callFramePrintState
	lastStepPrinted string
}

func newPrintState() *printState {
	return &printState{
		frames: map[string]callFramePrintState{},
	}
}

func (e *Emitter) Submit(run *v1.Run, prg *gptscript.Program, frames Frames) {
	e.liveStateLock.Lock()
	defer e.liveStateLock.Unlock()

	e.liveStates[run.Name] = append(e.liveStates[run.Name], liveState{Prg: prg, Frames: frames})
	e.liveBroadcast.Broadcast()
}

func (e *Emitter) Done(run *v1.Run) {
	e.liveStateLock.Lock()
	defer e.liveStateLock.Unlock()

	e.liveStates[run.Name] = append(e.liveStates[run.Name], liveState{Done: true})
	e.liveBroadcast.Broadcast()
}

func (e *Emitter) ClearProgress(run *v1.Run) {
	e.liveStateLock.Lock()
	defer e.liveStateLock.Unlock()

	delete(e.liveStates, run.Name)
	e.liveBroadcast.Broadcast()
}

func (e *Emitter) SubmitProgress(run *v1.Run, progress types.Progress) {
	e.liveStateLock.Lock()
	defer e.liveStateLock.Unlock()

	e.liveStates[run.Name] = append(e.liveStates[run.Name], liveState{Progress: &progress})
	e.liveBroadcast.Broadcast()
}

func (e *Emitter) findRunByThreadName(ctx context.Context, threadNamespace, threadName string) (*v1.Run, error) {
	var runs v1.RunList
	if err := e.client.List(ctx, &runs, kclient.InNamespace(threadNamespace)); err != nil {
		return nil, err
	}
	for _, run := range runs.Items {
		if run.Spec.ThreadName == threadName {
			return &run, nil
		}
	}
	return nil, fmt.Errorf("no run found for thread: %s", threadName)
}

func (e *Emitter) Watch(ctx context.Context, namespace string, opts WatchOptions) (chan types.Progress, error) {
	var (
		run v1.Run
	)

	if opts.Run != nil {
		run = *opts.Run
	} else if opts.LastRunName != "" {
		if err := e.client.Get(ctx, router.Key(namespace, opts.LastRunName), &run); err != nil {
			return nil, err
		}
	} else if opts.ThreadName != "" {
		var thread v1.Thread
		if err := e.client.Get(ctx, router.Key(namespace, opts.ThreadName), &thread); err != nil {
			return nil, err
		}
		if thread.Status.LastRunName == "" {
			runForThread, err := e.findRunByThreadName(ctx, namespace, opts.ThreadName)
			if err != nil {
				return nil, err
			}
			run = *runForThread
		} else if err := e.client.Get(ctx, router.Key(namespace, thread.Status.LastRunName), &run); err != nil {
			return nil, err
		}
	}

	result := make(chan types.Progress)

	if run.Name == "" {
		close(result)
		return result, nil
	}

	go func() {
		// error is ignored because it's internally sent to progress channel
		_ = e.streamEvents(ctx, run, opts, result)
	}()

	return result, nil
}

func (e *Emitter) printRun(ctx context.Context, state *printState, run v1.Run, result chan types.Progress) error {
	var (
		liveIndex    int
		broadcast    = make(chan struct{}, 1)
		done, cancel = context.WithCancel(ctx)
	)
	defer cancel()

	if run.Spec.WorkflowStepID != "" && run.Spec.WorkflowExecutionName != "" && state.lastStepPrinted != run.Spec.WorkflowStepID {
		var wfe v1.WorkflowExecution
		if err := e.client.Get(ctx, router.Key(run.Namespace, run.Spec.WorkflowExecutionName), &wfe); err != nil {
			return err
		}
		step := workflow.FindStep(wfe.Status.WorkflowManifest, run.Spec.WorkflowStepID)
		result <- types.Progress{
			RunID: run.Name,
			Step:  step,
		}
		state.lastStepPrinted = run.Spec.WorkflowStepID
	}

	go func() {
		e.liveStateLock.Lock()
		defer e.liveStateLock.Unlock()
		for {
			e.liveBroadcast.Wait()

			select {
			case broadcast <- struct{}{}:
			default:
			}

			select {
			case <-done.Done():
				return
			default:
			}
		}
	}()

	w, err := e.client.Watch(ctx, &v1.RunStateList{}, kclient.MatchingFields{"metadata.name": run.Name}, kclient.InNamespace(run.Namespace))
	if err != nil {
		return err
	}

	defer func() {
		if w != nil {
			w.Stop()
			for range w.ResultChan() {
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-broadcast:
			var notSeen []liveState
			e.liveStateLock.RLock()
			liveStateLen := len(e.liveStates[run.Name])
			if liveIndex < liveStateLen {
				notSeen = e.liveStates[run.Name][liveIndex:]
				liveIndex = liveStateLen
			}
			e.liveStateLock.RUnlock()
			if liveStateLen < liveIndex {
				return nil
			}
			for _, toPrint := range notSeen {
				if toPrint.Done {
					return nil
				}

				if toPrint.Progress != nil {
					result <- *toPrint.Progress
				} else {
					callToEvents(run.Name, toPrint.Prg, toPrint.Frames, state, result)
				}
			}
		case event, ok := <-w.ResultChan():
			if !ok {
				// resume
				w, err = e.client.Watch(ctx, &v1.RunStateList{}, kclient.MatchingFields{"metadata.name": run.Name}, kclient.InNamespace(run.Namespace))
				if err != nil {
					return err
				}
				continue
			}
			runState, ok := event.Object.(*v1.RunState)
			if !ok {
				continue
			}
			var (
				prg        gptscript.Program
				callFrames = Frames{}
			)
			if err := gz.Decompress(&prg, runState.Spec.Program); err != nil {
				return err
			}
			if err := gz.Decompress(&callFrames, runState.Spec.CallFrame); err != nil {
				return err
			}
			callToEvents(run.Name, &prg, callFrames, state, result)

			if runState.Spec.Done {
				if runState.Spec.Error != "" {
					return errors.New(runState.Spec.Error)
				}
				return nil
			}
		}
	}
}

func (e *Emitter) printParent(ctx context.Context, state *printState, run v1.Run, result chan types.Progress) error {
	if run.Spec.PreviousRunName == "" {
		return nil
	}

	var (
		parent      v1.Run
		errNotFound error
	)
	if err := e.client.Get(ctx, kclient.ObjectKey{Namespace: run.Namespace, Name: run.Spec.PreviousRunName}, &parent); err != nil {
		return err
	} else {
		if err := e.printParent(ctx, state, parent, result); apierrors.IsNotFound(err) {
			errNotFound = err
		} else if err != nil {
			return err
		}
	}

	return errors.Join(errNotFound, e.printRun(ctx, state, parent, result))
}

func (e *Emitter) streamEvents(ctx context.Context, run v1.Run, opts WatchOptions, result chan types.Progress) (retErr error) {
	defer close(result)
	defer func() {
		if retErr != nil {
			result <- types.Progress{
				RunID: run.Name,
				Error: retErr.Error(),
			}
		}
	}()

	for {
		state := newPrintState()

		if opts.History {
			if err := e.printParent(ctx, state, run, result); !apierrors.IsNotFound(err) && err != nil {
				return err
			}
		}

		if err := e.printRun(ctx, state, run, result); err != nil {
			return err
		}

		nextRun, err := e.findNextRun(ctx, run, opts.Follow)
		if err != nil {
			return err
		}
		if nextRun == nil {
			return nil
		}

		// don't tail history again
		opts.History = false
		run = *nextRun
	}
}

func (e *Emitter) isWorkflowDone(ctx context.Context, run v1.Run, follow bool) (chan struct{}, func(), error) {
	if run.Spec.WorkflowExecutionName == "" {
		return nil, func() {}, nil
	}
	w, err := e.client.Watch(ctx, &v1.WorkflowExecutionList{}, kclient.InNamespace(run.Namespace), &kclient.MatchingFields{
		"metadata.name": run.Spec.WorkflowExecutionName,
	})
	if err != nil {
		return nil, nil, err
	}

	result := make(chan struct{})
	cancel := func() {
		w.Stop()
		go func() {
			for range w.ResultChan() {
			}
		}()
	}

	go func() {
		defer cancel()
		for event := range w.ResultChan() {
			if wfe, ok := event.Object.(*v1.WorkflowExecution); ok {
				if wfe.Status.State == v1.WorkflowStateComplete || wfe.Status.State == v1.WorkflowStateError {
					close(result)
					return
				}
			}
		}
	}()

	return result, cancel, nil
}

func (e *Emitter) findNextRun(ctx context.Context, run v1.Run, follow bool) (*v1.Run, error) {
	var (
		runs     v1.RunList
		criteria = []kclient.ListOption{
			kclient.InNamespace(run.Namespace),
			kclient.MatchingLabels{v1.PreviousRunNameLabel: run.Name},
		}
	)

	if !follow && run.Spec.WorkflowExecutionName == "" {
		// If this isn't a workflow we are done at this point if follow is requested
		return nil, nil
	}

	if run.Spec.WorkflowExecutionName != "" && !follow {
		return nil, nil
	}

	if err := e.client.List(ctx, &runs, criteria...); err != nil {
		return nil, err
	}
	if len(runs.Items) > 0 {
		return &runs.Items[0], nil
	}
	w, err := e.client.Watch(ctx, &v1.RunList{}, append(criteria, &kclient.ListOptions{
		Raw: &metav1.ListOptions{
			ResourceVersion: runs.ResourceVersion,
			TimeoutSeconds:  typed.Pointer(int64(60 * 15)),
		},
	})...)
	if err != nil {
		return nil, err
	}
	defer func() {
		w.Stop()
		for range w.ResultChan() {
		}
	}()

	isWorkflowDone, cancel, err := e.isWorkflowDone(ctx, run, follow)
	if err != nil {
		return nil, err
	}
	defer cancel()

	for {
		select {
		case event, ok := <-w.ResultChan():
			if !ok {
				return nil, fmt.Errorf("failed to find next run after: %s", run.Name)
			}
			if run, ok := event.Object.(*v1.Run); ok {
				return run, nil
			}
		case <-isWorkflowDone:
			return nil, nil
		}
	}

}

func callToEvents(runID string, prg *gptscript.Program, frames Frames, printed *printState, out chan types.Progress) {
	var (
		parent gptscript.CallFrame
	)
	for _, frame := range frames {
		if frame.ParentID == "" {
			parent = frame
			break
		}
	}
	if parent.ID == "" {
		return
	}
	printCall(runID, prg, &parent, printed, out)
}

func printCall(runID string, prg *gptscript.Program, call *gptscript.CallFrame, lastPrint *printState, out chan types.Progress) {
	printed := lastPrint.frames[call.ID]
	lastOutputs := printed.Outputs

	if call.Input != "" && !printed.InputPrinted {
		out <- types.Progress{
			RunID:   runID,
			Content: "\n",
			Input:   call.Input,
		}
		printed.InputPrinted = true
	}

	for i, currentOutput := range call.Output {
		for i >= len(lastOutputs) {
			lastOutputs = append(lastOutputs, gptscript.Output{})
		}
		last := lastOutputs[i]

		if last.Content != currentOutput.Content {
			currentOutput.Content = printString(runID, out, last.Content, currentOutput.Content)
		}

		for _, callID := range slices.Sorted(maps.Keys(currentOutput.SubCalls)) {
			subCall := currentOutput.SubCalls[callID]
			if _, ok := last.SubCalls[callID]; !ok {
				if tool, ok := prg.ToolSet[subCall.ToolID]; ok {
					out <- types.Progress{
						RunID: runID,
						ToolCall: &types.ToolCall{
							Name:        tool.Name,
							Description: tool.Description,
							Input:       subCall.Input,
						},
					}
				}
			}
		}

		lastOutputs[i] = currentOutput
	}

	printed.Outputs = lastOutputs
	lastPrint.frames[call.ID] = printed
}

func printString(runID string, out chan types.Progress, last, current string) string {
	toPrint := current
	if strings.HasPrefix(current, last) {
		toPrint = current[len(last):]
	} else if len(last) > len(current) && strings.HasPrefix(last, current) {
		return last
	}

	var (
		toolName  string
		toolInput *types.ToolInput
	)

	toPrint, waitingOnModel := strings.CutPrefix(toPrint, "Waiting for model response...")
	toPrint, toolPrint, isToolCall := strings.Cut(toPrint, "<tool call> ")

	if isToolCall {
		toolName = strings.Split(toolPrint, " ->")[0]
	} else {
		_, wasToolPrint, wasToolCall := strings.Cut(current, "<tool call> ")
		if wasToolCall {
			toolName = strings.Split(wasToolPrint, " ->")[0]
			toolPrint = toPrint
			toPrint = ""
		}
	}

	toolPrint = strings.TrimPrefix(toolPrint, toolName+" -> ")

	if isToolCall {
		toolInput = &types.ToolInput{
			Content:          toolPrint,
			InternalToolName: toolName,
		}
	}

	out <- types.Progress{
		RunID:          runID,
		Content:        toPrint,
		ToolInput:      toolInput,
		WaitingOnModel: waitingOnModel,
	}

	return current
}
