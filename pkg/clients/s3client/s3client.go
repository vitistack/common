package s3client

// A generic S3 client implementation using MinIO SDK

import (
	"context"
	"io"
)

type S3Client interface {
	// Objects
	PutObject(ctx context.Context, objectName string, file io.Reader, size int64) error
	GetObject(ctx context.Context, objectName string) ([]byte, error)
	DeleteObject(ctx context.Context, objectName string) error
	ListObject(ctx context.Context, listOpt ListObjectsOptions) ([]string, error)
	// buckets
	CreateBucket(ctx context.Context) error
	DeleteBucket(ctx context.Context) error
}

type Options struct {
	Endpoint   string
	Region     string
	AccessKey  string
	SecretKey  string
	Secure     bool
	BucketName string
}

type Option func(*Options)

type ListObjectsOptions struct {
	Prefix    string
	Recursive bool
}

func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

func WithAccessKey(accessKey string) Option {
	return func(o *Options) {
		o.AccessKey = accessKey
	}
}

func WithSecretKey(secretKey string) Option {
	return func(o *Options) {
		o.SecretKey = secretKey
	}
}

func WithSecure(secure bool) Option {
	return func(o *Options) {
		o.Secure = secure
	}
}

func WithBucketName(bucketName string) Option {
	return func(o *Options) {
		o.BucketName = bucketName
	}
}

func WithRegion(region string) Option {
	return func(o *Options) {
		o.Region = region
	}
}

