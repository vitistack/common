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

func newS3Storage(cfg *Config) (ObjectStorage, error) {
	cfgCopy := *cfg
	if err := normalizeConfig(&cfgCopy); err != nil {
		return nil, err
	}
	cfg = &cfgCopy

	opts := []s3client.Option{
		s3client.WithRegion(cfg.Region),
		s3client.WithPathStyle(cfg.ForcePathStyle),
		s3client.WithInsecureSkipVerify(cfg.InsecureSkipTLSVerify),
	}

	// Only wire endpoint/SSL if user provided endpoint.
	if cfg.Endpoint != "" {
		endpoint, ssl, err := splitEndpoint(cfg.Endpoint)
		if err != nil {
			return nil, err
		}
		opts = append(opts,
			s3client.WithEndpoint(endpoint),
			s3client.WithSSL(ssl),
		)
	}

	accessKey := firstNonEmpty(os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("AWS_ACCESS_KEY_ID"))
	secretKey := firstNonEmpty(os.Getenv("S3_SECRET_ACCESS_KEY"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
	sessionTok := firstNonEmpty(os.Getenv("S3_SESSION_TOKEN"), os.Getenv("AWS_SESSION_TOKEN"))

	opts = append(opts, s3client.WithCredentials(accessKey, secretKey))
	if sessionTok != "" {
		opts = append(opts, s3client.WithSessionToken(sessionTok))
	}

	c, err := s3client.NewGenericS3Client(opts...)
	if err != nil {
		return nil, fmt.Errorf("object storage: init s3 client: %w", err)
	}

	return &s3Storage{cfg: *cfg, client: c}, nil
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
// Mutates cfg in-place to avoid copying a "heavy" Config (gocritic: hugeParam).
func normalizeConfig(cfg *Config) error {
	if cfg.Bucket == "" {
		return fmt.Errorf("object storage: bucket is required when enabled")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	// Endpoint optional: do NOT default to AWS.
	if cfg.Endpoint == "" {
		return nil
	}

	if !strings.HasPrefix(cfg.Endpoint, "http://") && !strings.HasPrefix(cfg.Endpoint, "https://") {
		return fmt.Errorf("object storage: endpoint must start with http:// or https:// (got %q)", cfg.Endpoint)
	}
	return nil
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
