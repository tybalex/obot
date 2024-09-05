package controller

import (
	"context"

	"github.com/acorn-io/baaah"
	"github.com/acorn-io/baaah/pkg/restconfig"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/storage/scheme"
)

type Services struct {
	AppName  string
	Router   *router.Router
	PreStart func(ctx context.Context) error
}

func NewServices(opt Options) (*Services, error) {
	apiServerRESTConfig, err := restconfig.FromURLTokenAndScheme(opt.ApiUrl, opt.ApiToken, scheme.Scheme)
	if err != nil {
		return nil, err
	}

	r, err := baaah.NewRouter("otto-controller", &baaah.Options{
		DefaultRESTConfig: apiServerRESTConfig,
		DefaultNamespace:  opt.Namespace,
		Scheme:            scheme.Scheme,
	})
	if err != nil {
		return nil, err
	}

	return &Services{
		AppName: opt.AppName,
		Router:  r,
		PreStart: func(ctx context.Context) error {
			return restconfig.WaitFor(ctx, apiServerRESTConfig)
		},
	}, nil
}
