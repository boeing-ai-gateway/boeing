//go:generate go run github.com/boeing-ai-gateway/nah/cmd/deepcopy ./pkg/storage/apis/boeing.boeing.ai/v1/
//go:generate go run github.com/boeing-ai-gateway/nah/cmd/deepcopy ./apiclient/types/
//go:generate go run k8s.io/kube-openapi/cmd/openapi-gen --go-header-file tools/header.txt --output-file openapi_generated.go --output-dir ./pkg/storage/openapi/generated/ --output-pkg github.com/boeing-ai-gateway/boeing/pkg/storage/openapi/generated github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1 k8s.io/apimachinery/pkg/apis/meta/v1 k8s.io/apimachinery/pkg/runtime k8s.io/apimachinery/pkg/version k8s.io/apimachinery/pkg/api/resource k8s.io/apimachinery/pkg/util/intstr k8s.io/api/coordination/v1 github.com/boeing-ai-gateway/boeing/apiclient/types

package main
