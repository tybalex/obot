//go:generate go run github.com/acorn-io/baaah/cmd/deepcopy ./pkg/storage/apis/otto.gptscript.ai/v1/
//go:generate go run github.com/acorn-io/baaah/cmd/deepcopy ./apiclient/types/
//go:generate go run k8s.io/kube-openapi/cmd/openapi-gen -i github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/version,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/util/intstr,k8s.io/api/coordination/v1,github.com/gptscript-ai/otto/apiclient/types -p ./pkg/storage/openapi/generated -h tools/header.txt

package main
