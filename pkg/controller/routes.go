package controller

import (
	"github.com/acorn-io/baaah/pkg/apply"
	"github.com/acorn-io/baaah/pkg/conditions"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/runs"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/threads"
	"github.com/gptscript-ai/otto/pkg/services"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func routes(router *router.Router, services *services.Services) error {
	root := router.Middleware(conditions.ErrorMiddleware())

	root.Type(&v1.Run{}).FinalizeFunc(v1.RunFinalizer, runs.DeleteRunState)
	root.Type(&v1.Run{}).HandlerFunc(runs.Cleanup)

	root.Type(&v1.Thread{}).HandlerFunc(threads.Cleanup)

	return nil
}

func gc(req router.Request, resp router.Response) error {
	return apply.New(req.Client).PurgeOrphan(req.Ctx, req.Object)
}
