package services

import (
	"github.com/acorn-io/baaah/pkg/randomtoken"
	"github.com/acorn-io/mink/pkg/db"
	"github.com/gptscript-ai/otto/pkg/storage/authn"
	"github.com/gptscript-ai/otto/pkg/storage/authz"
	"github.com/gptscript-ai/otto/pkg/storage/scheme"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

type Config struct {
	StorageListenPort int    `usage:"Port to storage backend will listen on (default: random port)"`
	StorageToken      string `usage:"Token for storage access, will be generated if not passed"`
	//AuditLogPath       string `usage:"Location of where to store audit logs"`
	//AuditLogPolicyFile string `usage:"Location of audit log policy file"`
	DSN           string `usage:"Database dsn in driver://connection_string format" default:"sqlite://file:otto.db?_journal=WAL&cache=shared&_busy_timeout=30000"`
	KnowledgeBin  string `usage:"Location of knowledge binary" default:"knowledge" env:"KNOWLEDGE_BIN"`
	KnowledgeTool string `usage:"The knowledge tool to use" default:"github.com/gptscript-ai/knowledge/gateway@v0.4.14-rc.2" env:"KNOWLEDGE_TOOL"`
}

type Services struct {
	DB    *db.Factory
	Authn authenticator.Request
	Authz authorizer.Authorizer
}

func New(config Config) (_ *Services, err error) {
	if config.StorageToken == "" {
		config.StorageToken, err = randomtoken.Generate()
		if err != nil {
			return nil, err
		}
	}

	dbClient, err := db.NewFactory(scheme.Scheme, config.DSN)
	if err != nil {
		return nil, err
	}

	services := &Services{
		DB:    dbClient,
		Authn: authn.NewAuthenticator(config.StorageToken),
		Authz: &authz.Authorizer{},
	}

	return services, nil
}
