// pkg/clients/objectstorage/smoke.go
package objectstorage

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

// SmokeUpload uploads a tiny object to verify connectivity/credentials.
// Intended for manual validation during bring-up.
func SmokeUpload(ctx context.Context, store ObjectStorage) error {
	if store == nil {
		return fmt.Errorf("object storage: smoke upload requested but store is nil (disabled?)")
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	key := "smoke-" + ts + ".txt"

	body := []byte("vitistack common objectstorage smoke test\n")
	return store.Put(ctx, key, bytes.NewReader(body), int64(len(body)), "text/plain")
}
