package s3mock

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/vitistack/common/pkg/clients/s3client/s3interface"
)

// MockS3Client is a mock implementation of the S3Client interface for testing
type MockS3Client struct {
	mu      sync.RWMutex
	objects map[string][]byte // objectName -> data

	// Track method calls for test verification
	PutObjectCalls    int
	GetObjectCalls    int
	DeleteObjectCalls int
	ListObjectCalls   int
	CreateBucketCalls int
	DeleteBucketCalls int

	// Error injection for testing error scenarios
	PutObjectErr    error
	GetObjectErr    error
	DeleteObjectErr error
	ListObjectErr   error
	CreateBucketErr error
	DeleteBucketErr error
}

// NewMockS3Client creates a new mock S3 client
func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		objects: make(map[string][]byte),
	}
}

// PutObject stores an object in memory
func (m *MockS3Client) PutObject(ctx context.Context, objectName string, file io.Reader, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PutObjectCalls++

	if m.PutObjectErr != nil {
		return m.PutObjectErr
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	m.objects[objectName] = data
	return nil
}

// GetObject retrieves an object from memory
func (m *MockS3Client) GetObject(ctx context.Context, objectName string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.GetObjectCalls++

	if m.GetObjectErr != nil {
		return nil, m.GetObjectErr
	}

	data, exists := m.objects[objectName]
	if !exists {
		return nil, fmt.Errorf("object not found: %s", objectName)
	}

	// Return a copy to prevent external modifications
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return dataCopy, nil
}

// DeleteObject removes an object from memory
func (m *MockS3Client) DeleteObject(ctx context.Context, objectName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteObjectCalls++

	if m.DeleteObjectErr != nil {
		return m.DeleteObjectErr
	}

	if _, exists := m.objects[objectName]; !exists {
		return fmt.Errorf("object not found: %s", objectName)
	}

	delete(m.objects, objectName)
	return nil
}

// ListObject returns a list of object names matching the criteria
func (m *MockS3Client) ListObject(ctx context.Context, listOpt s3interface.ListObjectsOptions) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.ListObjectCalls++

	if m.ListObjectErr != nil {
		return nil, m.ListObjectErr
	}

	var objects []string
	for key := range m.objects {
		// Filter by prefix if specified
		if listOpt.Prefix != "" {
			if len(key) >= len(listOpt.Prefix) && key[:len(listOpt.Prefix)] == listOpt.Prefix {
				objects = append(objects, key)
			}
		} else {
			objects = append(objects, key)
		}
	}

	return objects, nil
}

// CreateBucket is a no-op in the mock
func (m *MockS3Client) CreateBucket(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateBucketCalls++

	if m.CreateBucketErr != nil {
		return m.CreateBucketErr
	}

	return nil
}

// DeleteBucket is a no-op in the mock
func (m *MockS3Client) DeleteBucket(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteBucketCalls++

	if m.DeleteBucketErr != nil {
		return m.DeleteBucketErr
	}

	return nil
}

// Helper methods for testing

// Reset clears all stored data and resets call counters
func (m *MockS3Client) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.objects = make(map[string][]byte)
	m.PutObjectCalls = 0
	m.GetObjectCalls = 0
	m.DeleteObjectCalls = 0
	m.ListObjectCalls = 0
	m.CreateBucketCalls = 0
	m.DeleteBucketCalls = 0
	m.PutObjectErr = nil
	m.GetObjectErr = nil
	m.DeleteObjectErr = nil
	m.ListObjectErr = nil
	m.CreateBucketErr = nil
	m.DeleteBucketErr = nil
}

// SetObject directly sets an object (useful for test setup)
func (m *MockS3Client) SetObject(objectName string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[objectName] = data
}

// ObjectExists checks if an object exists
func (m *MockS3Client) ObjectExists(objectName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.objects[objectName]
	return exists
}

// ObjectCount returns the number of stored objects
func (m *MockS3Client) ObjectCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.objects)
}

// Ensure MockS3Client implements the S3Client interface
var _ s3interface.S3Client = (*MockS3Client)(nil)
