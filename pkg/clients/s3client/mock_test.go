package s3client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
)

func TestMockS3Client_CreateBucket(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Test creating a bucket
	err := client.CreateBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// Verify bucket exists
	exists, err := client.BucketExists(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("BucketExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected bucket to exist")
	}

	// Test creating duplicate bucket
	err = client.CreateBucket(ctx, "test-bucket")
	if err == nil {
		t.Error("Expected error when creating duplicate bucket")
	}
}

func TestMockS3Client_DeleteBucket(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Create and delete bucket
	_ = client.CreateBucket(ctx, "test-bucket")

	err := client.DeleteBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("DeleteBucket failed: %v", err)
	}

	// Verify bucket no longer exists
	exists, _ := client.BucketExists(ctx, "test-bucket")
	if exists {
		t.Error("Expected bucket to not exist")
	}

	// Test deleting non-existent bucket
	err = client.DeleteBucket(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent bucket")
	}
}

func TestMockS3Client_PutAndGetObject(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Create bucket first
	_ = client.CreateBucket(ctx, "test-bucket")

	// Test putting an object
	testData := []byte("hello, world!")
	reader := bytes.NewReader(testData)

	output, err := client.PutObject(ctx, "test-bucket", "test-key", reader, int64(len(testData)),
		WithContentType("text/plain"),
		WithMetadata(map[string]string{"custom": "value"}),
	)
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}
	if output.ETag == "" {
		t.Error("Expected ETag in output")
	}

	// Test getting the object
	getOutput, err := client.GetObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	defer func() { _ = getOutput.Body.Close() }()

	data, _ := io.ReadAll(getOutput.Body)
	if string(data) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, data)
	}
	if getOutput.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got %q", getOutput.ContentType)
	}
}

func TestMockS3Client_DeleteObject(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Setup
	_ = client.CreateBucket(ctx, "test-bucket")
	_, _ = client.PutObject(ctx, "test-bucket", "test-key", bytes.NewReader([]byte("data")), 4)

	// Delete object
	err := client.DeleteObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}

	// Verify object is gone
	_, err = client.GetObject(ctx, "test-bucket", "test-key")
	if err == nil {
		t.Error("Expected error when getting deleted object")
	}
}

func TestMockS3Client_ListObjects(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Setup
	_ = client.CreateBucket(ctx, "test-bucket")
	_, _ = client.PutObject(ctx, "test-bucket", "dir1/file1.txt", bytes.NewReader([]byte("data1")), 5)
	_, _ = client.PutObject(ctx, "test-bucket", "dir1/file2.txt", bytes.NewReader([]byte("data2")), 5)
	_, _ = client.PutObject(ctx, "test-bucket", "dir2/file3.txt", bytes.NewReader([]byte("data3")), 5)

	// List all objects
	output, err := client.ListObjects(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}
	if len(output.Objects) != 3 {
		t.Errorf("Expected 3 objects, got %d", len(output.Objects))
	}

	// List with prefix
	output, err = client.ListObjects(ctx, "test-bucket", WithPrefix("dir1/"))
	if err != nil {
		t.Fatalf("ListObjects with prefix failed: %v", err)
	}
	if len(output.Objects) != 2 {
		t.Errorf("Expected 2 objects with prefix 'dir1/', got %d", len(output.Objects))
	}
}

func TestMockS3Client_HeadObject(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Setup
	_ = client.CreateBucket(ctx, "test-bucket")
	_, _ = client.PutObject(ctx, "test-bucket", "test-key", bytes.NewReader([]byte("hello")), 5,
		WithContentType("text/plain"),
	)

	// Head object
	output, err := client.HeadObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("HeadObject failed: %v", err)
	}
	if output.ContentLength != 5 {
		t.Errorf("Expected content length 5, got %d", output.ContentLength)
	}
	if output.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got %q", output.ContentType)
	}
}

func TestMockS3Client_GetPresignedURL(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Setup
	_ = client.CreateBucket(ctx, "test-bucket")
	_, _ = client.PutObject(ctx, "test-bucket", "test-key", bytes.NewReader([]byte("hello")), 5)

	// Get presigned URL
	url, err := client.GetPresignedURL(ctx, "test-bucket", "test-key", 3600)
	if err != nil {
		t.Fatalf("GetPresignedURL failed: %v", err)
	}
	if url == "" {
		t.Error("Expected non-empty URL")
	}
}

func TestMockS3Client_Hooks(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Setup bucket
	_ = client.CreateBucket(ctx, "test-bucket")

	// Test PutObjectHook
	hookCalled := false
	client.PutObjectHook = func(ctx context.Context, bucket, key string) error {
		hookCalled = true
		return fmt.Errorf("simulated error")
	}

	_, err := client.PutObject(ctx, "test-bucket", "test-key", bytes.NewReader([]byte("data")), 4)
	if err == nil {
		t.Error("Expected error from hook")
	}
	if !hookCalled {
		t.Error("Expected hook to be called")
	}
}

func TestMockS3Client_Close(t *testing.T) {
	client := NewMockS3Client(
		WithEndpoint("mock.s3.local"),
		WithCredentials("test-key", "test-secret"),
	)

	ctx := context.Background()

	// Close the client
	_ = client.Close()

	// Operations should fail after close
	err := client.CreateBucket(ctx, "test-bucket")
	if err == nil {
		t.Error("Expected error after client is closed")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Endpoint:        "s3.amazonaws.com",
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: &Config{
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
			},
			wantErr: true,
		},
		{
			name: "missing access key",
			config: &Config{
				Endpoint:        "s3.amazonaws.com",
				SecretAccessKey: "secret-key",
			},
			wantErr: true,
		},
		{
			name: "missing secret key",
			config: &Config{
				Endpoint:    "s3.amazonaws.com",
				AccessKeyID: "access-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFunctionalOptions(t *testing.T) {
	cfg := DefaultConfig()

	// Apply options
	cfg.ApplyOptions(
		WithEndpoint("custom.endpoint.com"),
		WithRegion("eu-west-1"),
		WithCredentials("my-key", "my-secret"),
		WithSSL(false),
		WithPathStyle(true),
		WithMaxRetries(5),
		WithDebug(true),
	)

	if cfg.Endpoint != "custom.endpoint.com" {
		t.Errorf("Expected endpoint 'custom.endpoint.com', got %q", cfg.Endpoint)
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("Expected region 'eu-west-1', got %q", cfg.Region)
	}
	if cfg.AccessKeyID != "my-key" {
		t.Errorf("Expected access key 'my-key', got %q", cfg.AccessKeyID)
	}
	if cfg.SecretAccessKey != "my-secret" {
		t.Errorf("Expected secret key 'my-secret', got %q", cfg.SecretAccessKey)
	}
	if cfg.UseSSL != false {
		t.Error("Expected UseSSL to be false")
	}
	if cfg.PathStyle != true {
		t.Error("Expected PathStyle to be true")
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", cfg.MaxRetries)
	}
	if cfg.Debug != true {
		t.Error("Expected Debug to be true")
	}
}
