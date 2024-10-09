package webcontext

import (
	"context"

	"github.com/otto8-ai/otto8/apiclient"
)

type clientKey struct{}

func WithClient(ctx context.Context, client *apiclient.Client) context.Context {
	return context.WithValue(ctx, clientKey{}, client)
}

func AgentID(ctx context.Context) string {
	return "otto"
}

func ThreadID(ctx context.Context) string {
	return "t1-user"
}

func Client(ctx context.Context) *apiclient.Client {
	c, ok := ctx.Value(clientKey{}).(*apiclient.Client)
	if !ok {
		return nil
	}
	return c
}

type pageKey struct{}

func WithPage(ctx context.Context, page string) context.Context {
	return context.WithValue(ctx, pageKey{}, page)
}

func Page(ctx context.Context) string {
	p, ok := ctx.Value(pageKey{}).(string)
	if !ok {
		return "Chat"
	}
	return p
}
