package db

import (
	"fmt"

	"gorm.io/gorm"
)

func dropBoeingMCPTokensTable(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("boeing_mcp_tokens") {
		return nil
	}

	if err := tx.Migrator().DropTable("boeing_mcp_tokens"); err != nil {
		return fmt.Errorf("failed to drop boeing_mcp_tokens table: %w", err)
	}

	return nil
}
