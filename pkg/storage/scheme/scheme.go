package scheme

import (
	"github.com/acorn-io/baaah/pkg/restconfig"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
)

var (
	Scheme,
	Codecs,
	Parameter,
	AddToScheme = restconfig.MustBuildScheme(
		v1.AddToScheme,
		coordinationv1.AddToScheme,
		corev1.AddToScheme)
)
