// pkg/clients/objectstorage/s3_storage.go
package objectstorage

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Storage struct {
	cfg    Config
	client *s3.Client
}

func newS3Storage(cfg Config) (ObjectStorage, error) {
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	httpClient := awshttp.NewBuildableClient().WithTransportOptions(func(tr *http.Transport) {
		if cfg.InsecureSkipTLSVerify {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
		}
	})

	awsCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("object storage: load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &s3Storage{cfg: cfg, client: s3Client}, nil
}

func (s *s3Storage) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if key == "" {
		return fmt.Errorf("object storage: key is required")
	}

	fullKey := JoinKey(s.cfg.Prefix, key)

	in := &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(fullKey),
		Body:   body,
	}

	// Optional: helps some S3-compatible gateways; safe to omit when unknown.
	if size >= 0 {
		in.ContentLength = aws.Int64(size)
	}

	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}

	_, err := s.client.PutObject(ctx, in)
	if err != nil {
		return fmt.Errorf("object storage: put s3://%s/%s: %w", s.cfg.Bucket, fullKey, err)
	}
	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.Bucket == "" {
		return cfg, fmt.Errorf("object storage: bucket is required when enabled")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.Endpoint != "" && !(strings.HasPrefix(cfg.Endpoint, "http://") || strings.HasPrefix(cfg.Endpoint, "https://")) {
		return cfg, fmt.Errorf("object storage: endpoint must start with http:// or https:// (got %q)", cfg.Endpoint)
	}
	return cfg, nil
}
