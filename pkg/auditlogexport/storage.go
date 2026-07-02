package auditlogexport

import (
	"context"
	"fmt"
	"io"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
)

// StorageProvider defines the interface for all storage providers.
type StorageProvider interface {
	// Test tests if the storage provider is working.
	Test(ctx context.Context, config types.StorageConfig) error

	// Upload uploads the given data to the storage provider.
	Upload(ctx context.Context, config types.StorageConfig, bucket, key string, data io.Reader) error
}

// NewStorageProvider returns an error for all cloud storage provider types.
// External cloud storage export (S3, GCS, Azure) has been disabled.
func NewStorageProvider(providerType types.StorageProviderType) (StorageProvider, error) {
	return nil, fmt.Errorf("external storage provider %q is disabled: cloud export (S3, GCS, Azure) has been removed", providerType)
}
