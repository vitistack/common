package s3client

// A generic S3 client implementation using MinIO SDK

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/vitistack/common/pkg/loggers/vlog"
)

type S3Client interface {
	//Objects
	PutObject(ctx context.Context, objectName string, file io.Reader, size int64) error
	GetObject(ctx context.Context, objectName string) ([]byte, error)
	DeleteObject(ctx context.Context, objectName string) error
	ListObject(ctx context.Context, listOpt ListObjectsOptions) ([]string, error)

	//buckets
	CreateBucket(ctx context.Context) error
	DeleteBucket(ctx context.Context) error
	//BucketExists

	/*
	   //objects
	   GetObject
	   PutObject
	   listObjects
	   DeleteObjects

	   //buckets
	   CreateBucket
	   DeleteBucket
	   ListBuckets
	   modifybucket
	*/
}

type MinioS3Client struct {
	client     *minio.Client
	bucketName string
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

func NewS3Client(opt ...Option) (*MinioS3Client, error) {
	var options Options
	for _, o := range opt {
		o(&options)
	}
	m, err := minio.New(options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(options.AccessKey, options.SecretKey, ""),
		Secure: options.Secure,
		Region: options.Region,
	})

	if err != nil {
		vlog.Errorf("Failed to create S3 client: %v", err)
		return nil, err
	}

	s3client := &MinioS3Client{
		client:     m,
		bucketName: options.BucketName,
	}

	return s3client, nil
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

// implement struct for get options and get output
func (c *MinioS3Client) GetObject(ctx context.Context, objectName string) ([]byte, error) {

	object, err := c.client.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})

	if err != nil {
		vlog.Warnf("Failed to get object: %v", err)
		return nil, err
	}

	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		vlog.Errorf("Failed to read object data: %v", err)
		return nil, err
	}

	return data, nil
}

// implement struct for put options and put output
func (c *MinioS3Client) PutObject(ctx context.Context, objectName string, file io.Reader, size int64) error {

	_, err := c.client.PutObject(ctx, c.bucketName, objectName, file, size, minio.PutObjectOptions{})
	if err != nil {
		vlog.Warnf("Failed to put object: %v", err)
		return err
	}

	return nil
}

func (c *MinioS3Client) DeleteObject(ctx context.Context, objectName string) error {

	err := c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		vlog.Warnf("Failed to delete object: %v", err)
		return err
	}

	return nil
}

type ListObjectsOptions struct {
	Prefix    string
	Recursive bool
}

// implement struct for list options and list output
func (c *MinioS3Client) ListObject(ctx context.Context, listOpt ListObjectsOptions) ([]string, error) {

	objectCh := c.client.ListObjects(ctx, c.bucketName, minio.ListObjectsOptions{
		Prefix:    listOpt.Prefix,
		Recursive: listOpt.Recursive,
	})

	var objects []string
	for object := range objectCh {
		if object.Err != nil {
			vlog.Warnf("Failed to list objects: %v", object.Err)
			return nil, object.Err
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}

func (c *MinioS3Client) CreateBucket(ctx context.Context) error {

	err := c.client.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{})
	if err != nil {
		vlog.Warnf("Failed to create bucket: %v", err)
		return err
	}

	return nil
}

func (c *MinioS3Client) DeleteBucket(ctx context.Context) error {

	err := c.client.RemoveBucket(ctx, c.bucketName)
	if err != nil {
		vlog.Warnf("Failed to delete bucket: %v", err)
		return err
	}

	return nil
}
