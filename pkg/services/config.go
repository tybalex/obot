package services

import (
	"context"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/jwt"
	"github.com/gptscript-ai/otto/pkg/storage"
	"github.com/gptscript-ai/otto/pkg/storage/services"
	"github.com/gptscript-ai/otto/pkg/system"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
)

type Config struct {
	HTTPListenPort int `usage:"HTTP port to listen on" default:"8080" name:"http-listen-port"`
	services.Config
}

type Services struct {
	StorageClient   storage.Client
	GPTClient       *gptscript.GPTScript
	Invoker         *invoke.Invoker
	TokenServer     *jwt.TokenService
	APIServer       *api.Server
	WorkspaceClient *wclient.Client
}

func New(ctx context.Context, config Config) (*Services, error) {
	system.SetBinToSelf()

	storageClient, err := storage.Start(ctx, config.Config)
	if err != nil {
		return nil, err
	}

	c, err := gptscript.NewGPTScript()
	if err != nil {
		return nil, err
	}

	tokenServer := &jwt.TokenService{}

	workspaceClient := wclient.New()

	return &Services{
		StorageClient:   storageClient,
		GPTClient:       c,
		APIServer:       api.NewServer(storageClient, c, tokenServer),
		TokenServer:     tokenServer,
		WorkspaceClient: workspaceClient,
		Invoker:         invoke.NewInvoker(storageClient, c, tokenServer, workspaceClient),
	}, nil
}
