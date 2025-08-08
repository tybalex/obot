package mcpgateway

import (
	"context"
)

type contextKey struct{}

func withMessageContext(ctx context.Context, m messageContext) context.Context {
	return context.WithValue(ctx, contextKey{}, m)
}

func messageContextFromContext(ctx context.Context) (messageContext, bool) {
	m, ok := ctx.Value(contextKey{}).(messageContext)
	return m, ok
}
