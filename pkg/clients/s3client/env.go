package s3client

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variable names for S3 configuration
const (
	EnvS3Endpoint           = "S3_ENDPOINT"
	EnvS3Region             = "S3_REGION"
	EnvS3AccessKeyID        = "S3_ACCESS_KEY_ID"
	EnvS3SecretAccessKey    = "S3_SECRET_ACCESS_KEY"
	EnvS3SessionToken       = "S3_SESSION_TOKEN" // #nosec G101 -- This is an env var name, not a credential
	EnvS3UseSSL             = "S3_USE_SSL"
	EnvS3InsecureSkipVerify = "S3_INSECURE_SKIP_VERIFY"
	EnvS3ForceHTTP2         = "S3_FORCE_HTTP2"
	EnvS3PathStyle          = "S3_PATH_STYLE"
	EnvS3ConnectTimeout     = "S3_CONNECT_TIMEOUT"
	EnvS3RequestTimeout     = "S3_REQUEST_TIMEOUT"
	EnvS3MaxRetries         = "S3_MAX_RETRIES"
	EnvS3Debug              = "S3_DEBUG"
	EnvS3Bucket             = "S3_BUCKET"
	EnvS3Mock               = "S3_MOCK"
)

// ConfigFromEnv creates a Config populated from environment variables.
// Environment variables take precedence over defaults.
// Returns a Config that can be further customized with functional options.
//
// Supported environment variables:
//   - S3_ENDPOINT: S3 endpoint URL (required)
//   - S3_REGION: AWS region (default: "us-east-1")
//   - S3_ACCESS_KEY_ID: Access key for authentication (required)
//   - S3_SECRET_ACCESS_KEY: Secret key for authentication (required)
//   - S3_SESSION_TOKEN: Optional session token for temporary credentials
//   - S3_USE_SSL: Whether to use HTTPS (default: "true")
//   - S3_INSECURE_SKIP_VERIFY: Skip TLS certificate verification (default: "false")
//   - S3_FORCE_HTTP2: Enable HTTP/2 for connections (default: "false")
//   - S3_PATH_STYLE: Use path-style addressing (default: "false")
//   - S3_CONNECT_TIMEOUT: Connection timeout (default: "10s")
//   - S3_REQUEST_TIMEOUT: Request timeout (default: "30s")
//   - S3_MAX_RETRIES: Maximum retry attempts (default: "3")
//   - S3_DEBUG: Enable debug logging (default: "false")
func ConfigFromEnv() *Config {
	cfg := DefaultConfig()

	applyStringEnv(&cfg.Endpoint, EnvS3Endpoint)
	applyStringEnv(&cfg.Region, EnvS3Region)
	applyStringEnv(&cfg.AccessKeyID, EnvS3AccessKeyID)
	applyStringEnv(&cfg.SecretAccessKey, EnvS3SecretAccessKey)
	applyStringEnv(&cfg.SessionToken, EnvS3SessionToken)

	applyBoolEnv(&cfg.UseSSL, EnvS3UseSSL, true)
	applyBoolEnv(&cfg.InsecureSkipVerify, EnvS3InsecureSkipVerify, false)
	applyBoolEnv(&cfg.ForceHTTP2, EnvS3ForceHTTP2, false)
	applyBoolEnv(&cfg.PathStyle, EnvS3PathStyle, false)
	applyBoolEnv(&cfg.Debug, EnvS3Debug, false)

	applyDurationEnv(&cfg.ConnectTimeout, EnvS3ConnectTimeout)
	applyDurationEnv(&cfg.RequestTimeout, EnvS3RequestTimeout)
	applyIntEnv(&cfg.MaxRetries, EnvS3MaxRetries)

	return cfg
}

// applyStringEnv sets the target to the environment variable value if set.
func applyStringEnv(target *string, envVar string) {
	if val := os.Getenv(envVar); val != "" {
		*target = val
	}
}

// applyBoolEnv sets the target to the parsed boolean environment variable value if set.
func applyBoolEnv(target *bool, envVar string, defaultVal bool) {
	if val := os.Getenv(envVar); val != "" {
		*target = parseBool(val, defaultVal)
	}
}

// applyDurationEnv sets the target to the parsed duration environment variable value if valid.
func applyDurationEnv(target *time.Duration, envVar string) {
	if val := os.Getenv(envVar); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			*target = d
		}
	}
}

// applyIntEnv sets the target to the parsed integer environment variable value if valid.
func applyIntEnv(target *int, envVar string) {
	if val := os.Getenv(envVar); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			*target = n
		}
	}
}

// GetBucketFromEnv returns the bucket name from the S3_BUCKET environment variable.
func GetBucketFromEnv() string {
	return os.Getenv(EnvS3Bucket)
}

// NewMockS3ClientFromEnv creates a new MockS3Client configured from environment variables.
// Additional options can be passed to override environment settings.
func NewMockS3ClientFromEnv(opts ...Option) *MockS3Client {
	cfg := ConfigFromEnv()
	cfg.ApplyOptions(opts...)

	return &MockS3Client{
		buckets: make(map[string]*mockBucket),
		config:  cfg,
	}
}

// WithConfigFromEnv returns options that apply environment variable configuration.
// This can be used to mix env config with explicit options:
//
//	client := s3client.NewMockS3Client(
//	    s3client.WithConfigFromEnv(),
//	    s3client.WithDebug(true), // Override debug setting
//	)
func WithConfigFromEnv() Option {
	return func(c *Config) {
		envCfg := ConfigFromEnv()

		// Only apply non-empty values from env
		if envCfg.Endpoint != "" {
			c.Endpoint = envCfg.Endpoint
		}
		if envCfg.Region != "" {
			c.Region = envCfg.Region
		}
		if envCfg.AccessKeyID != "" {
			c.AccessKeyID = envCfg.AccessKeyID
		}
		if envCfg.SecretAccessKey != "" {
			c.SecretAccessKey = envCfg.SecretAccessKey
		}
		if envCfg.SessionToken != "" {
			c.SessionToken = envCfg.SessionToken
		}

		// Apply boolean and numeric values if explicitly set in env
		if os.Getenv(EnvS3UseSSL) != "" {
			c.UseSSL = envCfg.UseSSL
		}
		if os.Getenv(EnvS3InsecureSkipVerify) != "" {
			c.InsecureSkipVerify = envCfg.InsecureSkipVerify
		}
		if os.Getenv(EnvS3ForceHTTP2) != "" {
			c.ForceHTTP2 = envCfg.ForceHTTP2
		}
		if os.Getenv(EnvS3PathStyle) != "" {
			c.PathStyle = envCfg.PathStyle
		}
		if os.Getenv(EnvS3ConnectTimeout) != "" {
			c.ConnectTimeout = envCfg.ConnectTimeout
		}
		if os.Getenv(EnvS3RequestTimeout) != "" {
			c.RequestTimeout = envCfg.RequestTimeout
		}
		if os.Getenv(EnvS3MaxRetries) != "" {
			c.MaxRetries = envCfg.MaxRetries
		}
		if os.Getenv(EnvS3Debug) != "" {
			c.Debug = envCfg.Debug
		}
	}
}

// parseBool parses a string to boolean, returning defaultVal on error.
// Accepts: "true", "1", "yes", "on" as true values (case-insensitive).
func parseBool(s string, defaultVal bool) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultVal
	}
}
