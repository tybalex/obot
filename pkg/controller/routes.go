package controller

import (
	"github.com/acorn-io/baaah/pkg/apply"
	"github.com/acorn-io/baaah/pkg/conditions"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func routes(router *router.Router, services *Services) error {
	root := router.Middleware(conditions.ErrorMiddleware())
	root.Type(&v1.Agent{}).HandlerFunc(gc)

	return nil
}

func gc(req router.Request, resp router.Response) error {
	return apply.New(req.Client).PurgeOrphan(req.Ctx, req.Object)
}
