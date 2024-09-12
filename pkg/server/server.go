package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api/router"
	"github.com/gptscript-ai/otto/pkg/controller"
	"github.com/gptscript-ai/otto/pkg/mvl"
	"github.com/gptscript-ai/otto/pkg/services"
)

var log = mvl.Package()

func Run(ctx context.Context, c services.Config) error {
	svcs, err := services.New(ctx, c)
	if err != nil {
		return err
	}

	go func() {
		c, err := controller.New(ctx, svcs)
		if err != nil {
			log.Fatalf("Failed to start controller: %v", err)
		}
		if err := c.Start(ctx); err != nil {
			log.Fatalf("Failed to start controller: %v", err)
		}
	}()

	handler, err := router.Router(svcs)
	if err != nil {
		return err
	}

	context.AfterFunc(ctx, func() {
		log.Fatalf("Interrupted, exiting")
	})

	address := fmt.Sprintf("0.0.0.0:%d", c.HTTPListenPort)
	log.Infof("Starting server on %s", address)
	return http.ListenAndServe(address, handler)
}
