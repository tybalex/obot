package data

import (
	"context"
	_ "embed"

	"github.com/otto8-ai/nah/pkg/apply"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.otto8.ai/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

//go:embed otto.yaml
var ottoData []byte

func Data(ctx context.Context, c kclient.Client) error {
	var otto v1.Agent
	if err := yaml.Unmarshal(ottoData, &otto); err != nil {
		return err
	}

	return apply.Ensure(ctx, c, &otto)
}
