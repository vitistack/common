package objectstorage

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vitistack/common/pkg/clients/s3client"
)

func TestJoinKey(t *testing.T) {
	tests := []struct {
		prefix string
		key    string
		want   string
	}{
		{"", "a/b", "a/b"},
		{"p", "a/b", "p/a/b"},
		{"p/", "a/b", "p/a/b"},
		{"p///", "/a/b", "p/a/b"},
		{"p", "/a/b", "p/a/b"},
		{"p/", "/a/b", "p/a/b"},
	}

	for _, tt := range tests {
		got := JoinKey(tt.prefix, tt.key)
		if got != tt.want {
			t.Fatalf("JoinKey(%q, %q) = %q, want %q", tt.prefix, tt.key, got, tt.want)
		}
	}
}

func TestNormalizeConfig_Defaults(t *testing.T) {
	cfg := Config{
		Bucket: "b",
		Region: "",
		// Endpoint empty -> defaults to AWS endpoint in normalizeConfig
	}

	if err := normalizeConfig(&cfg); err != nil {
		t.Fatalf("normalizeConfig unexpected error: %v", err)
	}

	if cfg.Region != "us-east-1" {
		t.Fatalf("expected default region us-east-1, got %q", cfg.Region)
	}

	if cfg.Endpoint == "" {
		t.Fatalf("expected default endpoint to be set when empty")
	}
}
func TestNormalizeConfig_EndpointSchemeRequiredWhenProvided(t *testing.T) {
	cfg := Config{
		Bucket:   "b",
		Region:   "us-east-1",
		Endpoint: "s3.example.com",
	}

	err := normalizeConfig(&cfg)
	if err == nil {
		t.Fatalf("expected error when endpoint has no scheme")
	}
}

func TestSplitEndpoint(t *testing.T) {
	host, ssl, err := splitEndpoint("http://localhost:4566")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if host != "localhost:4566" || ssl != false {
		t.Fatalf("got (%q,%v), want (%q,%v)", host, ssl, "localhost:4566", false)
	}

	host, ssl, err = splitEndpoint("https://s3.amazonaws.com")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if host != "s3.amazonaws.com" || ssl != true {
		t.Fatalf("got (%q,%v), want (%q,%v)", host, ssl, "s3.amazonaws.com", true)
	}
}

func TestPut_UsesPrefixAndWrapsClient(t *testing.T) {
	ctx := context.Background()

	mock := s3client.NewMockS3Client()
	if err := mock.CreateBucket(ctx, "bucket"); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	s := &s3Storage{
		cfg: Config{
			Bucket: "bucket",
			Prefix: "vitistack",
		},
		client: mock,
	}

	body := []byte("hello")
	if err := s.Put(ctx, "snapshots/a.txt", bytes.NewReader(body), int64(len(body)), "text/plain"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Verify stored under prefix/key (mock stores by bucket+objectKey)
	_, err := mock.GetObject(ctx, "bucket", "vitistack/snapshots/a.txt")
	if err != nil {
		t.Fatalf("expected object to exist at vitistack/snapshots/a.txt: %v", err)
	}
}

func TestPut_ErrorWrapping(t *testing.T) {
	ctx := context.Background()

	mock := s3client.NewMockS3Client()
	_ = mock.CreateBucket(ctx, "bucket")

	wantErr := errors.New("boom")
	mock.PutObjectHook = func(ctx context.Context, bucketName, objectKey string) error {
		return wantErr
	}

	s := &s3Storage{
		cfg: Config{
			Bucket: "bucket",
			Prefix: "vitistack",
		},
		client: mock,
	}

	err := s.Put(ctx, "x.txt", strings.NewReader("x"), 1, "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "object storage: put s3://bucket/vitistack/x.txt") {
		t.Fatalf("expected wrapped path in error, got: %v", err)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected errors.Is(err, wantErr)=true, got: %v", err)
	}
}
