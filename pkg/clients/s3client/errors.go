// Package s3client provides common S3 error types.
package s3client

import (
	"errors"
	"fmt"
)

// S3Error represents an S3-specific error.
type S3Error struct {
	Code       string
	Message    string
	StatusCode int
	RequestID  string
}

func (e *S3Error) Error() string {
	return fmt.Sprintf("S3Error: %s - %s (StatusCode: %d, RequestID: %s)",
		e.Code, e.Message, e.StatusCode, e.RequestID)
}

// Common S3 error codes
const (
	ErrCodeNoSuchBucket          = "NoSuchBucket"
	ErrCodeNoSuchKey             = "NoSuchKey"
	ErrCodeBucketAlreadyExists   = "BucketAlreadyExists"
	ErrCodeBucketNotEmpty        = "BucketNotEmpty"
	ErrCodeAccessDenied          = "AccessDenied"
	ErrCodeInvalidAccessKeyID    = "InvalidAccessKeyId"
	ErrCodeSignatureDoesNotMatch = "SignatureDoesNotMatch"
	ErrCodeRequestTimeout        = "RequestTimeout"
	ErrCodeInternalError         = "InternalError"
)

// NewS3Error creates a new S3Error.
func NewS3Error(code, message string, statusCode int) *S3Error {
	return &S3Error{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// IsNotFoundError checks if the error is a "not found" error.
func IsNotFoundError(err error) bool {
	var s3Err *S3Error
	if errors.As(err, &s3Err) {
		return s3Err.Code == ErrCodeNoSuchBucket || s3Err.Code == ErrCodeNoSuchKey
	}
	return false
}

// IsAccessDeniedError checks if the error is an access denied error.
func IsAccessDeniedError(err error) bool {
	var s3Err *S3Error
	if errors.As(err, &s3Err) {
		return s3Err.Code == ErrCodeAccessDenied
	}
	return false
}

// IsBucketExistsError checks if the error is a bucket already exists error.
func IsBucketExistsError(err error) bool {
	var s3Err *S3Error
	if errors.As(err, &s3Err) {
		return s3Err.Code == ErrCodeBucketAlreadyExists
	}
	return false
}
