// pkg/clients/objectstorage/object_storage_service.go
package objectstorage

import (
	"context"
	"io"
)

type Config struct {
	Enabled               bool
	Endpoint              string // optional for S3-compatible backends
	Region                string
	Bucket                string
	Prefix                string
	ForcePathStyle        bool
	InsecureSkipTLSVerify bool
}

// ObjectStorage is a minimal interface for uploading backup artifacts.
// Add Get/List/Delete later when backup/restore/retention needs it.
type ObjectStorage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
}

// New returns an ObjectStorage implementation based on config.
// For now, return nil when disabled; caller should handle nil safely.
func New(cfg Config) (ObjectStorage, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	return newS3Storage(cfg)
}

// helper to build keys consistently
func JoinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	// avoid double slashes
	for len(prefix) > 0 && prefix[len(prefix)-1] == '/' {
		prefix = prefix[:len(prefix)-1]
	}
	for len(key) > 0 && key[0] == '/' {
		key = key[1:]
	}
	return prefix + "/" + key
}
