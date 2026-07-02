//nolint:revive
package api

import (
	"net/http"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
)

var ErrMustAuth = &types.ErrHTTP{
	Code:    http.StatusUnauthorized,
	Message: "unauthorized request, must authenticate",
}
