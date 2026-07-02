//nolint:revive
package types

import (
	"time"

	types2 "github.com/boeing-ai-gateway/boeing/apiclient/types"
)

type TempSetupUser struct {
	ID                    uint        `json:"id" gorm:"primaryKey"`
	UserID                uint        `json:"userID" gorm:"index"`
	Username              string      `json:"username"`
	Email                 string      `json:"email"`
	Role                  types2.Role `json:"role"`
	IconURL               string      `json:"iconURL"`
	AuthProviderName      string      `json:"authProviderName"`
	AuthProviderNamespace string      `json:"authProviderNamespace"`
	CreatedAt             time.Time   `json:"createdAt"`
}
