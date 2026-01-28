package s3client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"
)

// MockS3Client is a mock implementation of S3Client for testing purposes.
// It stores objects in memory and simulates S3 operations.
type MockS3Client struct {
	mu      sync.RWMutex
	buckets map[string]*mockBucket
	config  *Config
	closed  bool

	// Hooks for testing specific scenarios
	PutObjectHook    func(ctx context.Context, bucket, key string) error
	GetObjectHook    func(ctx context.Context, bucket, key string) error
	DeleteObjectHook func(ctx context.Context, bucket, key string) error
	ListObjectsHook  func(ctx context.Context, bucket string) error
}

type mockBucket struct {
	objects map[string]*mockObject
	created time.Time
}

type mockObject struct {
	data         []byte
	contentType  string
	metadata     map[string]string
	lastModified time.Time
	etag         string
	storageClass string
}

// NewMockS3Client creates a new mock S3 client for testing.
func NewMockS3Client(opts ...Option) *MockS3Client {
	cfg := DefaultConfig()
	cfg.ApplyOptions(opts...)

	return &MockS3Client{
		buckets: make(map[string]*mockBucket),
		config:  cfg,
	}
}

// PutObject stores an object in memory.
func (m *MockS3Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts ...PutObjectOption) (*PutObjectOutput, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	if m.PutObjectHook != nil {
		if err := m.PutObjectHook(ctx, bucket, key); err != nil {
			return nil, err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket %q does not exist", bucket)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	options := &PutObjectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	etag := fmt.Sprintf("%x", len(data))
	obj := &mockObject{
		data:         data,
		contentType:  options.ContentType,
		metadata:     options.Metadata,
		lastModified: time.Now(),
		etag:         etag,
		storageClass: options.StorageClass,
	}

	b.objects[key] = obj

	return &PutObjectOutput{
		ETag: etag,
	}, nil
}

// GetObject retrieves an object from memory.
func (m *MockS3Client) GetObject(ctx context.Context, bucket, key string, opts ...GetObjectOption) (*GetObjectOutput, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	if m.GetObjectHook != nil {
		if err := m.GetObjectHook(ctx, bucket, key); err != nil {
			return nil, err
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.buckets[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket %q does not exist", bucket)
	}

	obj, ok := b.objects[key]
	if !ok {
		return nil, fmt.Errorf("object %q does not exist in bucket %q", key, bucket)
	}

	return &GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(obj.data)),
		ContentType:   obj.contentType,
		ContentLength: int64(len(obj.data)),
		ETag:          obj.etag,
		LastModified:  obj.lastModified,
		Metadata:      obj.metadata,
	}, nil
}

// DeleteObject removes an object from memory.
func (m *MockS3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	if m.DeleteObjectHook != nil {
		if err := m.DeleteObjectHook(ctx, bucket, key); err != nil {
			return err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[bucket]
	if !ok {
		return fmt.Errorf("bucket %q does not exist", bucket)
	}

	delete(b.objects, key)
	return nil
}

// ListObjects lists objects in the bucket.
func (m *MockS3Client) ListObjects(ctx context.Context, bucket string, opts ...ListObjectsOption) (*ListObjectsOutput, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	if m.ListObjectsHook != nil {
		if err := m.ListObjectsHook(ctx, bucket); err != nil {
			return nil, err
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.buckets[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket %q does not exist", bucket)
	}

	options := &ListObjectsOptions{
		MaxKeys: 1000,
	}
	for _, opt := range opts {
		opt(options)
	}

	var objects []Object
	prefixSet := make(map[string]struct{})

	// Collect and sort keys for deterministic output
	keys := make([]string, 0, len(b.objects))
	for key := range b.objects {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		obj := b.objects[key]
		// Apply prefix filter
		if options.Prefix != "" && !hasPrefix(key, options.Prefix) {
			continue
		}

		// Handle delimiter for virtual directories
		if options.Delimiter != "" {
			suffix := key[len(options.Prefix):]
			if idx := indexOf(suffix, options.Delimiter); idx >= 0 {
				prefix := options.Prefix + suffix[:idx+1]
				prefixSet[prefix] = struct{}{}
				continue
			}
		}

		objects = append(objects, Object{
			Key:          key,
			Size:         int64(len(obj.data)),
			ETag:         obj.etag,
			LastModified: obj.lastModified,
			StorageClass: obj.storageClass,
		})

		if options.MaxKeys > 0 && len(objects) >= int(options.MaxKeys) {
			break
		}
	}

	prefixes := make([]string, 0, len(prefixSet))
	for p := range prefixSet {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	return &ListObjectsOutput{
		Objects:     objects,
		Prefixes:    prefixes,
		IsTruncated: false,
	}, nil
}

// HeadObject retrieves metadata about an object.
func (m *MockS3Client) HeadObject(ctx context.Context, bucket, key string) (*HeadObjectOutput, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.buckets[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket %q does not exist", bucket)
	}

	obj, ok := b.objects[key]
	if !ok {
		return nil, fmt.Errorf("object %q does not exist in bucket %q", key, bucket)
	}

	return &HeadObjectOutput{
		ContentType:   obj.contentType,
		ContentLength: int64(len(obj.data)),
		ETag:          obj.etag,
		LastModified:  obj.lastModified,
		Metadata:      obj.metadata,
	}, nil
}

// BucketExists checks if a bucket exists.
func (m *MockS3Client) BucketExists(ctx context.Context, bucket string) (bool, error) {
	if m.closed {
		return false, fmt.Errorf("client is closed")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.buckets[bucket]
	return ok, nil
}

// CreateBucket creates a new bucket.
func (m *MockS3Client) CreateBucket(ctx context.Context, bucket string, opts ...CreateBucketOption) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.buckets[bucket]; ok {
		return fmt.Errorf("bucket %q already exists", bucket)
	}

	m.buckets[bucket] = &mockBucket{
		objects: make(map[string]*mockObject),
		created: time.Now(),
	}
	return nil
}

// DeleteBucket deletes an empty bucket.
func (m *MockS3Client) DeleteBucket(ctx context.Context, bucket string) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[bucket]
	if !ok {
		return fmt.Errorf("bucket %q does not exist", bucket)
	}

	if len(b.objects) > 0 {
		return fmt.Errorf("bucket %q is not empty", bucket)
	}

	delete(m.buckets, bucket)
	return nil
}

// GetPresignedURL generates a mock presigned URL.
func (m *MockS3Client) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	if m.closed {
		return "", fmt.Errorf("client is closed")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.buckets[bucket]
	if !ok {
		return "", fmt.Errorf("bucket %q does not exist", bucket)
	}

	if _, ok := b.objects[key]; !ok {
		return "", fmt.Errorf("object %q does not exist in bucket %q", key, bucket)
	}

	// Generate a mock presigned URL
	expiry := time.Now().Add(expires).Unix()
	return fmt.Sprintf("https://%s/%s/%s?expires=%d&signature=mock", m.config.Endpoint, bucket, key, expiry), nil
}

// Close marks the client as closed.
func (m *MockS3Client) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// Reset clears all buckets and objects (useful for test setup/teardown).
func (m *MockS3Client) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buckets = make(map[string]*mockBucket)
	m.closed = false
}

// GetBucketCount returns the number of buckets (for testing).
func (m *MockS3Client) GetBucketCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.buckets)
}

// GetObjectCount returns the number of objects in a bucket (for testing).
func (m *MockS3Client) GetObjectCount(bucket string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if b, ok := m.buckets[bucket]; ok {
		return len(b.objects)
	}
	return 0
}

// Helper functions

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Ensure MockS3Client implements S3Client interface
var _ S3Client = (*MockS3Client)(nil)
