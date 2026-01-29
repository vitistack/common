// Package s3client provides a general S3 bucket client with support for
// multiple S3-compatible backends (AWS S3, MinIO, etc.) using the
// Golang Functional Options Pattern.
package s3client

import (
	"context"
	"fmt"
	"io"
	"time"
)

// S3Client defines the interface for S3 bucket operations.
// This interface can be implemented by different S3-compatible backends.
type S3Client interface {
	// PutObject uploads an object to the specified bucket.
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts ...PutObjectOption) (*PutObjectOutput, error)

	// GetObject retrieves an object from the specified bucket.
	GetObject(ctx context.Context, bucket, key string, opts ...GetObjectOption) (*GetObjectOutput, error)

	// DeleteObject deletes an object from the specified bucket.
	DeleteObject(ctx context.Context, bucket, key string) error

	// ListObjects lists objects in the specified bucket with optional prefix.
	ListObjects(ctx context.Context, bucket string, opts ...ListObjectsOption) (*ListObjectsOutput, error)

	// HeadObject retrieves metadata about an object without returning the object itself.
	HeadObject(ctx context.Context, bucket, key string) (*HeadObjectOutput, error)

	// BucketExists checks if a bucket exists.
	BucketExists(ctx context.Context, bucket string) (bool, error)

	// CreateBucket creates a new bucket.
	CreateBucket(ctx context.Context, bucket string, opts ...CreateBucketOption) error

	// DeleteBucket deletes an empty bucket.
	DeleteBucket(ctx context.Context, bucket string) error

	// GetPresignedURL generates a presigned URL for the object.
	GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)

	// Close closes the client and releases any resources.
	Close() error
}

// Object represents an S3 object in a listing.
type Object struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	StorageClass string
}

// PutObjectOutput contains the result of a PutObject operation.
type PutObjectOutput struct {
	ETag      string
	VersionID string
}

// GetObjectOutput contains the result of a GetObject operation.
type GetObjectOutput struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	Metadata      map[string]string
}

// ListObjectsOutput contains the result of a ListObjects operation.
type ListObjectsOutput struct {
	Objects               []Object
	Prefixes              []string
	IsTruncated           bool
	NextContinuationToken string
}

// HeadObjectOutput contains metadata about an object.
type HeadObjectOutput struct {
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	Metadata      map[string]string
	VersionID     string
}

// Config holds the configuration for creating an S3 client.
type Config struct {
	// Endpoint is the S3 endpoint URL (e.g., "s3.amazonaws.com" or "minio.example.com:9000")
	Endpoint string

	// Region is the AWS region (e.g., "us-east-1")
	Region string

	// AccessKeyID is the access key for authentication
	AccessKeyID string

	// SecretAccessKey is the secret key for authentication
	SecretAccessKey string

	// SessionToken is an optional session token for temporary credentials
	SessionToken string

	// UseSSL determines whether to use HTTPS
	UseSSL bool

	// InsecureSkipVerify skips TLS certificate verification (use with caution)
	InsecureSkipVerify bool

	// ForceHTTP2 enables HTTP/2 for connections (default: false for better compatibility)
	ForceHTTP2 bool

	// PathStyle forces path-style addressing (required for MinIO and some S3-compatible services)
	PathStyle bool

	// ConnectTimeout is the timeout for establishing connections
	ConnectTimeout time.Duration

	// RequestTimeout is the timeout for individual requests
	RequestTimeout time.Duration

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int

	// Debug enables debug logging
	Debug bool
}

// Option is a functional option for configuring the S3 client.
type Option func(*Config)

// WithEndpoint sets the S3 endpoint.
func WithEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}

// WithRegion sets the AWS region.
func WithRegion(region string) Option {
	return func(c *Config) {
		c.Region = region
	}
}

// WithCredentials sets the access key and secret key.
func WithCredentials(accessKeyID, secretAccessKey string) Option {
	return func(c *Config) {
		c.AccessKeyID = accessKeyID
		c.SecretAccessKey = secretAccessKey
	}
}

// WithSessionToken sets the session token for temporary credentials.
func WithSessionToken(token string) Option {
	return func(c *Config) {
		c.SessionToken = token
	}
}

// WithSSL enables or disables SSL/TLS.
func WithSSL(useSSL bool) Option {
	return func(c *Config) {
		c.UseSSL = useSSL
	}
}

// WithInsecureSkipVerify skips TLS certificate verification.
// WARNING: This should only be used for testing or in environments with self-signed certificates.
func WithInsecureSkipVerify(skip bool) Option {
	return func(c *Config) {
		c.InsecureSkipVerify = skip
	}
}

// WithForceHTTP2 enables or disables HTTP/2 for connections.
// Default is false for better compatibility with various S3 providers.
func WithForceHTTP2(force bool) Option {
	return func(c *Config) {
		c.ForceHTTP2 = force
	}
}

// WithPathStyle enables or disables path-style addressing.
func WithPathStyle(pathStyle bool) Option {
	return func(c *Config) {
		c.PathStyle = pathStyle
	}
}

// WithConnectTimeout sets the connection timeout.
func WithConnectTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = timeout
	}
}

// WithRequestTimeout sets the request timeout.
func WithRequestTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.RequestTimeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithDebug enables or disables debug logging.
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Region:             "no-west-1",
		UseSSL:             true,
		InsecureSkipVerify: false,
		ForceHTTP2:         false,
		PathStyle:          false,
		ConnectTimeout:     10 * time.Second,
		RequestTimeout:     30 * time.Second,
		MaxRetries:         3,
		Debug:              false,
	}
}

// ApplyOptions applies the given options to the config.
func (c *Config) ApplyOptions(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if c.AccessKeyID == "" {
		return fmt.Errorf("access key ID is required")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("secret access key is required")
	}
	return nil
}

// PutObjectOptions holds options for the PutObject operation.
type PutObjectOptions struct {
	ContentType        string
	ContentEncoding    string
	ContentDisposition string
	CacheControl       string
	Metadata           map[string]string
	ACL                string
	StorageClass       string
}

// PutObjectOption is a functional option for PutObject.
type PutObjectOption func(*PutObjectOptions)

// WithContentType sets the content type for the object.
func WithContentType(contentType string) PutObjectOption {
	return func(o *PutObjectOptions) {
		o.ContentType = contentType
	}
}

// WithContentEncoding sets the content encoding for the object.
func WithContentEncoding(encoding string) PutObjectOption {
	return func(o *PutObjectOptions) {
		o.ContentEncoding = encoding
	}
}

// WithContentDisposition sets the content disposition for the object.
func WithContentDisposition(disposition string) PutObjectOption {
	return func(o *PutObjectOptions) {
		o.ContentDisposition = disposition
	}
}

// WithCacheControl sets the cache control header for the object.
func WithCacheControl(cacheControl string) PutObjectOption {
	return func(o *PutObjectOptions) {
		o.CacheControl = cacheControl
	}
}

// WithMetadata sets custom metadata for the object.
func WithMetadata(metadata map[string]string) PutObjectOption {
	return func(o *PutObjectOptions) {
		o.Metadata = metadata
	}
}

// WithACL sets the ACL for the object.
func WithACL(acl string) PutObjectOption {
	return func(o *PutObjectOptions) {
		o.ACL = acl
	}
}

// WithStorageClass sets the storage class for the object.
func WithStorageClass(storageClass string) PutObjectOption {
	return func(o *PutObjectOptions) {
		o.StorageClass = storageClass
	}
}

// GetObjectOptions holds options for the GetObject operation.
type GetObjectOptions struct {
	Range     string
	VersionID string
}

// GetObjectOption is a functional option for GetObject.
type GetObjectOption func(*GetObjectOptions)

// WithRange sets the byte range for partial object retrieval.
func WithRange(rangeHeader string) GetObjectOption {
	return func(o *GetObjectOptions) {
		o.Range = rangeHeader
	}
}

// WithVersionID sets the version ID for versioned objects.
func WithVersionID(versionID string) GetObjectOption {
	return func(o *GetObjectOptions) {
		o.VersionID = versionID
	}
}

// ListObjectsOptions holds options for the ListObjects operation.
type ListObjectsOptions struct {
	Prefix            string
	Delimiter         string
	MaxKeys           int32
	ContinuationToken string
}

// ListObjectsOption is a functional option for ListObjects.
type ListObjectsOption func(*ListObjectsOptions)

// WithPrefix sets the prefix filter for listing objects.
func WithPrefix(prefix string) ListObjectsOption {
	return func(o *ListObjectsOptions) {
		o.Prefix = prefix
	}
}

// WithDelimiter sets the delimiter for grouping objects.
func WithDelimiter(delimiter string) ListObjectsOption {
	return func(o *ListObjectsOptions) {
		o.Delimiter = delimiter
	}
}

// WithMaxKeys sets the maximum number of keys to return.
func WithMaxKeys(maxKeys int32) ListObjectsOption {
	return func(o *ListObjectsOptions) {
		o.MaxKeys = maxKeys
	}
}

// WithContinuationToken sets the continuation token for pagination.
func WithContinuationToken(token string) ListObjectsOption {
	return func(o *ListObjectsOptions) {
		o.ContinuationToken = token
	}
}

// CreateBucketOptions holds options for the CreateBucket operation.
type CreateBucketOptions struct {
	ACL                string
	ObjectLocking      bool
	LocationConstraint string
}

// CreateBucketOption is a functional option for CreateBucket.
type CreateBucketOption func(*CreateBucketOptions)

// WithBucketACL sets the ACL for the bucket.
func WithBucketACL(acl string) CreateBucketOption {
	return func(o *CreateBucketOptions) {
		o.ACL = acl
	}
}

// WithObjectLocking enables object locking for the bucket.
func WithObjectLocking(enabled bool) CreateBucketOption {
	return func(o *CreateBucketOptions) {
		o.ObjectLocking = enabled
	}
}

// WithLocationConstraint sets the location constraint for the bucket.
func WithLocationConstraint(location string) CreateBucketOption {
	return func(o *CreateBucketOptions) {
		o.LocationConstraint = location
	}
}
