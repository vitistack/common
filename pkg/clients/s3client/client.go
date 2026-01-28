package s3client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// GenericS3Client is a generic S3 client implementation using the MinIO SDK.
// It works with any S3-compatible storage provider:
//   - AWS S3
//   - MinIO
//   - Hetzner Object Storage
//   - DigitalOcean Spaces
//   - Wasabi
//   - Backblaze B2
//   - Google Cloud Storage (S3 compatibility)
//   - Azure Blob Storage (via S3 gateway)
//   - Cloudflare R2
//   - And many more...
type GenericS3Client struct {
	client *minio.Client
	config *Config
}

// NewGenericS3Client creates a new generic S3 client that works with any S3-compatible storage.
func NewGenericS3Client(opts ...Option) (*GenericS3Client, error) {
	cfg := DefaultConfig()
	cfg.ApplyOptions(opts...)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Parse endpoint to extract host (remove protocol if present)
	endpoint := cfg.Endpoint
	if u, err := url.Parse(cfg.Endpoint); err == nil && u.Host != "" {
		endpoint = u.Host
	}

	// Create MinIO client options
	minioOpts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
		Secure: cfg.UseSSL,
	}

	// Create the MinIO client
	client, err := minio.New(endpoint, minioOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return &GenericS3Client{
		client: client,
		config: cfg,
	}, nil
}

// NewGenericS3ClientFromEnv creates a new generic S3 client configured from environment variables.
func NewGenericS3ClientFromEnv(opts ...Option) (*GenericS3Client, error) {
	allOpts := append([]Option{WithConfigFromEnv()}, opts...)
	return NewGenericS3Client(allOpts...)
}

// PutObject uploads an object to the specified bucket.
func (c *GenericS3Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts ...PutObjectOption) (*PutObjectOutput, error) {
	options := &PutObjectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	putOpts := minio.PutObjectOptions{
		ContentType:        options.ContentType,
		ContentEncoding:    options.ContentEncoding,
		ContentDisposition: options.ContentDisposition,
		CacheControl:       options.CacheControl,
		UserMetadata:       options.Metadata,
		StorageClass:       options.StorageClass,
	}

	info, err := c.client.PutObject(ctx, bucket, key, reader, size, putOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	return &PutObjectOutput{
		ETag:      info.ETag,
		VersionID: info.VersionID,
	}, nil
}

// GetObject retrieves an object from the specified bucket.
func (c *GenericS3Client) GetObject(ctx context.Context, bucket, key string, opts ...GetObjectOption) (*GetObjectOutput, error) {
	options := &GetObjectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	getOpts := minio.GetObjectOptions{}
	if options.VersionID != "" {
		getOpts.VersionID = options.VersionID
	}
	// Note: Range header support would require parsing "bytes=0-1023" format
	// and using getOpts.SetRange(). Skipped for simplicity.

	obj, err := c.client.GetObject(ctx, bucket, key, getOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Get object info for metadata
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &GetObjectOutput{
		Body:          obj,
		ContentType:   info.ContentType,
		ContentLength: info.Size,
		ETag:          info.ETag,
		LastModified:  info.LastModified,
		Metadata:      info.UserMetadata,
	}, nil
}

// DeleteObject deletes an object from the specified bucket.
func (c *GenericS3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	err := c.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// DeleteObjectVersioned deletes a specific version of an object.
func (c *GenericS3Client) DeleteObjectVersioned(ctx context.Context, bucket, key, versionID string) error {
	err := c.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{
		VersionID: versionID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete object version: %w", err)
	}
	return nil
}

// ListObjects lists objects in the specified bucket with optional prefix.
func (c *GenericS3Client) ListObjects(ctx context.Context, bucket string, opts ...ListObjectsOption) (*ListObjectsOutput, error) {
	options := &ListObjectsOptions{
		MaxKeys: 1000,
	}
	for _, opt := range opts {
		opt(options)
	}

	listOpts := minio.ListObjectsOptions{
		Prefix:    options.Prefix,
		Recursive: options.Delimiter == "", // If no delimiter, list recursively
		MaxKeys:   int(options.MaxKeys),
	}

	output := &ListObjectsOutput{
		Objects:  make([]Object, 0),
		Prefixes: make([]string, 0),
	}

	// Use channel-based listing
	objectCh := c.client.ListObjects(ctx, bucket, listOpts)

	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", obj.Err)
		}

		// Check if this is a "directory" (common prefix)
		if options.Delimiter != "" && len(obj.Key) > 0 && obj.Key[len(obj.Key)-1] == '/' {
			output.Prefixes = append(output.Prefixes, obj.Key)
		} else {
			output.Objects = append(output.Objects, Object{
				Key:          obj.Key,
				Size:         obj.Size,
				ETag:         obj.ETag,
				LastModified: obj.LastModified,
				StorageClass: obj.StorageClass,
			})
		}
	}

	return output, nil
}

// HeadObject retrieves metadata about an object without returning the object itself.
func (c *GenericS3Client) HeadObject(ctx context.Context, bucket, key string) (*HeadObjectOutput, error) {
	info, err := c.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to head object: %w", err)
	}

	return &HeadObjectOutput{
		ContentType:   info.ContentType,
		ContentLength: info.Size,
		ETag:          info.ETag,
		LastModified:  info.LastModified,
		Metadata:      info.UserMetadata,
		VersionID:     info.VersionID,
	}, nil
}

// BucketExists checks if a bucket exists.
func (c *GenericS3Client) BucketExists(ctx context.Context, bucket string) (bool, error) {
	exists, err := c.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket: %w", err)
	}
	return exists, nil
}

// CreateBucket creates a new bucket.
func (c *GenericS3Client) CreateBucket(ctx context.Context, bucket string, opts ...CreateBucketOption) error {
	options := &CreateBucketOptions{}
	for _, opt := range opts {
		opt(options)
	}

	makeOpts := minio.MakeBucketOptions{
		ObjectLocking: options.ObjectLocking,
	}

	if options.LocationConstraint != "" {
		makeOpts.Region = options.LocationConstraint
	} else if c.config.Region != "" {
		makeOpts.Region = c.config.Region
	}

	err := c.client.MakeBucket(ctx, bucket, makeOpts)
	if err != nil {
		// Check if bucket already exists
		exists, existsErr := c.client.BucketExists(ctx, bucket)
		if existsErr == nil && exists {
			return fmt.Errorf("bucket %q already exists", bucket)
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// DeleteBucket deletes an empty bucket.
func (c *GenericS3Client) DeleteBucket(ctx context.Context, bucket string) error {
	err := c.client.RemoveBucket(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}

// GetPresignedURL generates a presigned URL for downloading an object.
func (c *GenericS3Client) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, bucket, key, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// GetPresignedPutURL generates a presigned URL for uploading an object.
func (c *GenericS3Client) GetPresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(ctx, bucket, key, expires)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned put URL: %w", err)
	}
	return presignedURL.String(), nil
}

// Close closes the client and releases any resources.
func (c *GenericS3Client) Close() error {
	// MinIO client doesn't require explicit cleanup
	return nil
}

// CopyObject copies an object from one location to another.
func (c *GenericS3Client) CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	src := minio.CopySrcOptions{
		Bucket: srcBucket,
		Object: srcKey,
	}
	dst := minio.CopyDestOptions{
		Bucket: dstBucket,
		Object: dstKey,
	}

	_, err := c.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}
	return nil
}

// ListBuckets lists all buckets.
func (c *GenericS3Client) ListBuckets(ctx context.Context) ([]string, error) {
	buckets, err := c.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	result := make([]string, len(buckets))
	for i, b := range buckets {
		result[i] = b.Name
	}
	return result, nil
}

// FPutObject uploads a file to the bucket.
func (c *GenericS3Client) FPutObject(ctx context.Context, bucket, key, filePath string, opts ...PutObjectOption) (*PutObjectOutput, error) {
	options := &PutObjectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	putOpts := minio.PutObjectOptions{
		ContentType:        options.ContentType,
		ContentEncoding:    options.ContentEncoding,
		ContentDisposition: options.ContentDisposition,
		CacheControl:       options.CacheControl,
		UserMetadata:       options.Metadata,
		StorageClass:       options.StorageClass,
	}

	info, err := c.client.FPutObject(ctx, bucket, key, filePath, putOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &PutObjectOutput{
		ETag:      info.ETag,
		VersionID: info.VersionID,
	}, nil
}

// FGetObject downloads an object to a file.
func (c *GenericS3Client) FGetObject(ctx context.Context, bucket, key, filePath string) error {
	err := c.client.FGetObject(ctx, bucket, key, filePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

// GetUnderlyingClient returns the underlying MinIO client for advanced operations.
func (c *GenericS3Client) GetUnderlyingClient() *minio.Client {
	return c.client
}

// Ensure GenericS3Client implements S3Client interface
var _ S3Client = (*GenericS3Client)(nil)
