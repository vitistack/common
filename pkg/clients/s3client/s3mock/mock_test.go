package s3mock

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/vitistack/common/pkg/clients/s3client/s3interface"
)

func TestNewMockS3Client(t *testing.T) {
	mock := NewMockS3Client()
	if mock == nil {
		t.Fatal("NewMockS3Client returned nil")
	}
	if mock.objects == nil {
		t.Fatal("objects map not initialized")
	}
}

func TestMockS3Client_PutObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	testData := []byte("test data")
	objectName := "test.txt"

	err := mock.PutObject(ctx, objectName, bytes.NewReader(testData), int64(len(testData)))
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// Verify the object was stored
	if !mock.ObjectExists(objectName) {
		t.Error("Object should exist after PutObject")
	}

	// Verify stored data
	stored, exists := mock.objects[objectName]
	if !exists {
		t.Fatal("Object not found in storage")
	}
	if !bytes.Equal(stored, testData) {
		t.Errorf("Stored data mismatch. Expected %v, got %v", testData, stored)
	}
}

func TestMockS3Client_GetObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	testData := []byte("hello world")
	objectName := "greeting.txt"

	// Setup: store an object
	mock.SetObject(objectName, testData)

	// Get the object
	data, err := mock.GetObject(ctx, objectName)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	// Verify data
	if !bytes.Equal(data, testData) {
		t.Errorf("Expected %s, got %s", testData, data)
	}
}

func TestMockS3Client_GetObject_NotFound(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	_, err := mock.GetObject(ctx, "non-existent.txt")
	if err == nil {
		t.Error("Expected error for non-existent object, got nil")
	}
}

func TestMockS3Client_DeleteObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	objectName := "to-delete.txt"
	mock.SetObject(objectName, []byte("data"))

	// Verify object exists before deletion
	if !mock.ObjectExists(objectName) {
		t.Fatal("Object should exist before deletion")
	}

	// Delete the object
	err := mock.DeleteObject(ctx, objectName)
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}

	// Verify object no longer exists
	if mock.ObjectExists(objectName) {
		t.Error("Object should not exist after deletion")
	}
}

func TestMockS3Client_DeleteObject_NotFound(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	err := mock.DeleteObject(ctx, "non-existent.txt")
	if err == nil {
		t.Error("Expected error when deleting non-existent object, got nil")
	}
}

func TestMockS3Client_ListObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	// Setup: add multiple objects
	mock.SetObject("file1.txt", []byte("data1"))
	mock.SetObject("file2.txt", []byte("data2"))
	mock.SetObject("dir/file3.txt", []byte("data3"))

	// Test: list all objects
	objects, err := mock.ListObject(ctx, s3interface.ListObjectsOptions{})
	if err != nil {
		t.Fatalf("ListObject failed: %v", err)
	}

	if len(objects) != 3 {
		t.Errorf("Expected 3 objects, got %d", len(objects))
	}
}

func TestMockS3Client_ListObject_WithPrefix(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	// Setup: add objects with different prefixes
	mock.SetObject("logs/2024/app.log", []byte("log1"))
	mock.SetObject("logs/2025/app.log", []byte("log2"))
	mock.SetObject("backups/db.sql", []byte("backup"))
	mock.SetObject("config.json", []byte("config"))

	// Test: list with prefix
	objects, err := mock.ListObject(ctx, s3interface.ListObjectsOptions{Prefix: "logs/"})
	if err != nil {
		t.Fatalf("ListObject with prefix failed: %v", err)
	}

	if len(objects) != 2 {
		t.Errorf("Expected 2 objects with prefix 'logs/', got %d", len(objects))
	}

	// Verify only log files are returned
	for _, obj := range objects {
		if len(obj) < 5 || obj[:5] != "logs/" {
			t.Errorf("Object %s does not match prefix 'logs/'", obj)
		}
	}
}

func TestMockS3Client_CreateBucket(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	err := mock.CreateBucket(ctx)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
}

func TestMockS3Client_DeleteBucket(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	err := mock.DeleteBucket(ctx)
	if err != nil {
		t.Fatalf("DeleteBucket failed: %v", err)
	}
}

func TestMockS3Client_ErrorInjection_PutObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	expectedErr := errors.New("storage full")
	mock.PutObjectErr = expectedErr

	err := mock.PutObject(ctx, "test.txt", bytes.NewReader([]byte("data")), 4)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockS3Client_ErrorInjection_GetObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	expectedErr := errors.New("network error")
	mock.GetObjectErr = expectedErr

	_, err := mock.GetObject(ctx, "test.txt")
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockS3Client_ErrorInjection_DeleteObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	expectedErr := errors.New("permission denied")
	mock.DeleteObjectErr = expectedErr

	err := mock.DeleteObject(ctx, "test.txt")
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockS3Client_ErrorInjection_ListObject(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	expectedErr := errors.New("timeout")
	mock.ListObjectErr = expectedErr

	_, err := mock.ListObject(ctx, s3interface.ListObjectsOptions{})
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockS3Client_Reset(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	// Populate with data and calls
	mock.SetObject("test.txt", []byte("data"))
	_ = mock.PutObject(ctx, "another.txt", bytes.NewReader([]byte("data")), 4)
	_, _ = mock.GetObject(ctx, "test.txt")
	mock.PutObjectErr = errors.New("some error")

	// Verify state before reset
	if mock.ObjectCount() != 2 {
		t.Errorf("Expected 2 objects before reset, got %d", mock.ObjectCount())
	}

	// Reset
	mock.Reset()

	// Verify state after reset
	if mock.ObjectCount() != 0 {
		t.Errorf("Expected 0 objects after reset, got %d", mock.ObjectCount())
	}
	if mock.PutObjectErr != nil {
		t.Error("Expected error to be cleared after reset")
	}
}

func TestMockS3Client_ObjectCount(t *testing.T) {
	mock := NewMockS3Client()

	if mock.ObjectCount() != 0 {
		t.Errorf("Expected 0 objects initially, got %d", mock.ObjectCount())
	}

	mock.SetObject("file1.txt", []byte("data1"))
	mock.SetObject("file2.txt", []byte("data2"))

	if mock.ObjectCount() != 2 {
		t.Errorf("Expected 2 objects, got %d", mock.ObjectCount())
	}
}

func TestMockS3Client_ConcurrentAccess(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	// Test concurrent writes
	done := make(chan bool)
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			objectName := bytes.NewBufferString("file")
			objectName.WriteString(string(rune('0' + (index % 10))))
			objectName.WriteString(".txt")

			data := []byte("data")
			_ = mock.PutObject(ctx, objectName.String(), bytes.NewReader(data), int64(len(data)))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify no race conditions occurred
	if mock.ObjectCount() == 0 {
		t.Error("Expected objects to be created")
	}
}

func TestMockS3Client_GetObject_ReturnsCopy(t *testing.T) {
	mock := NewMockS3Client()
	ctx := context.Background()

	originalData := []byte("original data")
	objectName := "test.txt"
	mock.SetObject(objectName, originalData)

	// Get the object
	data1, err := mock.GetObject(ctx, objectName)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	// Modify the returned data
	data1[0] = 'X'

	// Get the object again
	data2, err := mock.GetObject(ctx, objectName)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	// Verify the original data is unchanged
	if !bytes.Equal(data2, originalData) {
		t.Error("GetObject should return a copy, not a reference to the original data")
	}
}
