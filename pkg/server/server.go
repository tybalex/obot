package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gptscript-ai/gptscript/pkg/mvl"
	"github.com/gptscript-ai/otto/pkg/api/router"
	"github.com/gptscript-ai/otto/pkg/services"
)

var log = mvl.Package()

func Run(ctx context.Context, c services.Config) error {
	svcs, err := services.New(ctx, c)
	if err != nil {
		return err
	}

	handler, err := router.Router(svcs)
	if err != nil {
		return err
	}

	address := fmt.Sprintf("0.0.0.0:%d", c.HTTPListenPort)
	log.Infof("Starting server on %s", address)
	return http.ListenAndServe(address, handler)
}
