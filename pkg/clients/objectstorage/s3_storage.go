// pkg/clients/objectstorage/s3_storage.go
package objectstorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vitistack/common/pkg/clients/s3client"
)

type s3Storage struct {
	cfg    Config
	client s3client.S3Client
}

var _ ObjectStorage = (*s3Storage)(nil)

func newS3Storage(cfg Config) (ObjectStorage, error) {
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Decide SSL + strip scheme for s3client/minio-style endpoints.
	endpoint, ssl, err := splitEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	opts := []s3client.Option{
		s3client.WithEndpoint(endpoint),
		s3client.WithSSL(ssl),
		s3client.WithRegion(cfg.Region),
		s3client.WithPathStyle(cfg.ForcePathStyle),
		s3client.WithInsecureSkipVerify(cfg.InsecureSkipTLSVerify),
	}

	// Credentials: prefer S3_* but allow AWS_* as fallback.
	accessKey := firstNonEmpty(os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("AWS_ACCESS_KEY_ID"))
	secretKey := firstNonEmpty(os.Getenv("S3_SECRET_ACCESS_KEY"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
	sessionTok := firstNonEmpty(os.Getenv("S3_SESSION_TOKEN"), os.Getenv("AWS_SESSION_TOKEN"))

	// s3client expects creds to exist (no IAM role support).
	opts = append(opts, s3client.WithCredentials(accessKey, secretKey))
	if sessionTok != "" {
		opts = append(opts, s3client.WithSessionToken(sessionTok))
	}

	c, err := s3client.NewGenericS3Client(opts...)
	if err != nil {
		return nil, fmt.Errorf("object storage: init s3 client: %w", err)
	}

	return &s3Storage{cfg: cfg, client: c}, nil
}

func (s *s3Storage) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if key == "" {
		return fmt.Errorf("object storage: key is required")
	}

	fullKey := JoinKey(s.cfg.Prefix, key)

	var putOpts []s3client.PutObjectOption
	if contentType != "" {
		putOpts = append(putOpts, s3client.WithContentType(contentType))
	}

	_, err := s.client.PutObject(ctx, s.cfg.Bucket, fullKey, body, size, putOpts...)
	if err != nil {
		return fmt.Errorf("object storage: put s3://%s/%s: %w", s.cfg.Bucket, fullKey, err)
	}
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// normalizeConfig keeps objectstorage behavior predictable and testable.
func normalizeConfig(cfg Config) (Config, error) {
	if cfg.Bucket == "" {
		return cfg, fmt.Errorf("object storage: bucket is required when enabled")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	// Endpoint is optional: default to AWS S3.
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://s3.amazonaws.com"
		return cfg, nil
	}

	// If user sets endpoint, require explicit scheme so we can reliably decide SSL.
	if !(strings.HasPrefix(cfg.Endpoint, "http://") || strings.HasPrefix(cfg.Endpoint, "https://")) {
		return cfg, fmt.Errorf("object storage: endpoint must start with http:// or https:// (got %q)", cfg.Endpoint)
	}

	return cfg, nil
}

// splitEndpoint converts "http(s)://host[:port]" -> ("host[:port]", sslBool)
func splitEndpoint(endpoint string) (string, bool, error) {
	if strings.HasPrefix(endpoint, "http://") {
		return strings.TrimPrefix(endpoint, "http://"), false, nil
	}
	if strings.HasPrefix(endpoint, "https://") {
		return strings.TrimPrefix(endpoint, "https://"), true, nil
	}
	// normalizeConfig should prevent this except for defaulting logic changes later.
	return "", false, fmt.Errorf("object storage: endpoint missing scheme: %q", endpoint)
}
