package invoke

import (
	"context"
	"encoding/json"

	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.gptscript.ai/v1"
)

type SystemTaskOptions struct {
	CredentialContextIDs []string
}

func complete(opts []SystemTaskOptions) (result SystemTaskOptions) {
	for _, opt := range opts {
		result.CredentialContextIDs = append(result.CredentialContextIDs, opt.CredentialContextIDs...)
	}
	return
}

func (i *Invoker) SystemTask(ctx context.Context, thread *v1.Thread, tool, input any, opts ...SystemTaskOptions) (*Response, error) {
	opt := complete(opts)

	var inputString string
	switch v := input.(type) {
	case string:
		inputString = v
	case []byte:
		inputString = string(v)
	case nil:
		inputString = ""
	default:
		data, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		inputString = string(data)
	}
	// dumb hack to catch nil pointers than might be a nil value in a non-nil interface
	if inputString == "null" {
		inputString = ""
	}

	return i.createRun(ctx, i.uncached, thread, tool, inputString, runOptions{
		CredentialContextIDs: opt.CredentialContextIDs,
		Synchronous:          true,
	})
}
