package scheme

import (
	"github.com/boeing-ai-gateway/nah/pkg/restconfig"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
)

//nolint:revive
var Scheme, Codecs, Parameter, AddToScheme = restconfig.MustBuildScheme(
	v1.AddToScheme,
	coordinationv1.AddToScheme,
	corev1.AddToScheme,
)
