package events

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/gz"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
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
	Progress *v1.Progress
	Done     bool
}

type WatchOptions struct {
	History     bool
	LastRunName string
	Run         *v1.Run
}

type Frames map[string]gptscript.CallFrame

type printState map[string][]gptscript.Output

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

func (e *Emitter) SubmitProgress(run *v1.Run, progress v1.Progress) {
	e.liveStateLock.Lock()
	defer e.liveStateLock.Unlock()

	e.liveStates[run.Name] = append(e.liveStates[run.Name], liveState{Progress: &progress})
	e.liveBroadcast.Broadcast()
}

func (e *Emitter) Watch(ctx context.Context, thread *v1.Thread, opts WatchOptions) (chan v1.Progress, error) {
	var (
		run v1.Run
	)

	if opts.Run != nil {
		run = *opts.Run
	} else if opts.LastRunName != "" {
		if err := e.client.Get(ctx, router.Key(thread.Namespace, opts.LastRunName), &run); err != nil {
			return nil, err
		}
	} else {
		if err := e.client.Get(ctx, router.Key(thread.Namespace, thread.Status.LastRunName), &run); err != nil {
			return nil, err
		}
	}

	result := make(chan v1.Progress)

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

func (e *Emitter) printRun(ctx context.Context, state printState, run v1.Run, result chan v1.Progress) error {
	var (
		liveIndex    int
		broadcast    = make(chan struct{}, 1)
		done, cancel = context.WithCancel(ctx)
	)
	defer cancel()

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
					e.callsToEvents(toPrint.Prg, toPrint.Frames, state, result)
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
			e.callsToEvents(&prg, callFrames, state, result)

			if runState.Spec.Done {
				if runState.Spec.Error != "" {
					return errors.New(runState.Spec.Error)
				}
				return nil
			}
		}
	}
}

func (e *Emitter) printParent(ctx context.Context, state printState, run v1.Run, result chan v1.Progress) error {
	if run.Spec.PreviousRunName == "" {
		return nil
	}

	var parent v1.Run
	if err := e.client.Get(ctx, kclient.ObjectKey{Namespace: run.Namespace, Name: run.Spec.PreviousRunName}, &parent); err != nil {
		if err := e.printParent(ctx, state, parent, result); err != nil {
			return err
		}
	}

	return e.printRun(ctx, state, parent, result)
}

func (e *Emitter) streamEvents(ctx context.Context, run v1.Run, opts WatchOptions, result chan v1.Progress) (retErr error) {
	defer close(result)
	defer func() {
		if retErr != nil {
			result <- v1.Progress{Error: retErr.Error()}
		}
	}()

	state := printState{}

	if opts.History {
		if err := e.printParent(ctx, state, run, result); err != nil {
			return
		}
	}

	return e.printRun(ctx, state, run, result)
}

func (e *Emitter) callsToEvents(prg *gptscript.Program, frames Frames, printed printState, out chan v1.Progress) {
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
	printCall(prg, &parent, printed, out)
}

func printCall(prg *gptscript.Program, call *gptscript.CallFrame, lastPrint map[string][]gptscript.Output, out chan v1.Progress) {
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

		for _, callID := range slices.Sorted(maps.Keys(currentOutput.SubCalls)) {
			subCall := currentOutput.SubCalls[callID]
			if _, ok := last.SubCalls[callID]; !ok {
				if tool, ok := prg.ToolSet[subCall.ToolID]; ok {
					out <- v1.Progress{
						Tool: v1.ToolProgress{
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

	lastPrint[call.ID] = lastOutputs
}

func printString(out chan v1.Progress, last, current string) {
	toPrint := current
	if strings.HasPrefix(current, last) {
		toPrint = current[len(last):]
	}

	toPrint, waitingOnModel := strings.CutPrefix(toPrint, "Waiting for model response...")
	toPrint, toolPrint, isToolCall := strings.Cut(toPrint, "<tool call> ")
	toolName := ""

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

	out <- v1.Progress{
		Content: toPrint,
		Tool: v1.ToolProgress{
			GeneratingInputForName: toolName,
			GeneratingInput:        isToolCall,
			PartialInput:           toolPrint,
		},
		WaitingOnModel: waitingOnModel,
	}
}
