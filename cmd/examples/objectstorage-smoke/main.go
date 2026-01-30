// cmd/examples/objectstorage-smoke/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/vitistack/common/pkg/clients/objectstorage"
)

func main() {
	cfg := objectstorage.Config{
		Enabled:               true,
		Endpoint:              os.Getenv("S3_ENDPOINT"),
		Region:                getenvDefault("S3_REGION", "us-east-1"),
		Bucket:                os.Getenv("S3_BUCKET"),
		Prefix:                getenvDefault("S3_PREFIX", "vitistack"),
		ForcePathStyle:        getenvBoolDefault("S3_FORCE_PATH_STYLE", true),
		InsecureSkipTLSVerify: getenvBoolDefault("S3_INSECURE_SKIP_TLS_VERIFY", false),
	}

	if cfg.Bucket == "" {
		_, _ = fmt.Fprintln(os.Stderr, "missing S3_BUCKET")
		return
	}

	store, err := objectstorage.New(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "init objectstorage: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := objectstorage.SmokeUpload(ctx, store); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "smoke upload failed: %v\n", err)
		return
	}

	_, _ = fmt.Fprintln(os.Stdout, "smoke upload ok")
}
func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvBoolDefault(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	case "0", "false", "FALSE", "no", "NO":
		return false
	default:
		return def
	}
}
