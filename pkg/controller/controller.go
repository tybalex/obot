package controller

import (
	"context"
	"fmt"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/services"
	// Enabled logrus logging in baaah
	_ "github.com/acorn-io/baaah/pkg/logrus"
)

type Controller struct {
	router   *router.Router
	services *services.Services
}

func New(ctx context.Context, services *services.Services) (*Controller, error) {
	err := routes(services.Router, services)
	if err != nil {
		return nil, err
	}

	return &Controller{
		router:   services.Router,
		services: services,
	}, nil
}

func (c *Controller) Start(ctx context.Context) error {
	if err := c.router.Start(ctx); err != nil {
		return fmt.Errorf("failed to start router: %w", err)
	}
	return nil
}
