package s3client

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variable names for S3 configuration
const (
	EnvS3Endpoint        = "S3_ENDPOINT"
	EnvS3Region          = "S3_REGION"
	EnvS3AccessKeyID     = "S3_ACCESS_KEY_ID"
	EnvS3SecretAccessKey = "S3_SECRET_ACCESS_KEY"
	EnvS3SessionToken    = "S3_SESSION_TOKEN" // #nosec G101 -- This is an env var name, not a credential
	EnvS3UseSSL          = "S3_USE_SSL"
	EnvS3PathStyle       = "S3_PATH_STYLE"
	EnvS3ConnectTimeout  = "S3_CONNECT_TIMEOUT"
	EnvS3RequestTimeout  = "S3_REQUEST_TIMEOUT"
	EnvS3MaxRetries      = "S3_MAX_RETRIES"
	EnvS3Debug           = "S3_DEBUG"
	EnvS3Bucket          = "S3_BUCKET"
	EnvS3Mock            = "S3_MOCK"
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
//   - S3_PATH_STYLE: Use path-style addressing (default: "false")
//   - S3_CONNECT_TIMEOUT: Connection timeout (default: "10s")
//   - S3_REQUEST_TIMEOUT: Request timeout (default: "30s")
//   - S3_MAX_RETRIES: Maximum retry attempts (default: "3")
//   - S3_DEBUG: Enable debug logging (default: "false")
func ConfigFromEnv() *Config {
	cfg := DefaultConfig()

	if endpoint := os.Getenv(EnvS3Endpoint); endpoint != "" {
		cfg.Endpoint = endpoint
	}

	if region := os.Getenv(EnvS3Region); region != "" {
		cfg.Region = region
	}

	if accessKey := os.Getenv(EnvS3AccessKeyID); accessKey != "" {
		cfg.AccessKeyID = accessKey
	}

	if secretKey := os.Getenv(EnvS3SecretAccessKey); secretKey != "" {
		cfg.SecretAccessKey = secretKey
	}

	if sessionToken := os.Getenv(EnvS3SessionToken); sessionToken != "" {
		cfg.SessionToken = sessionToken
	}

	if useSSL := os.Getenv(EnvS3UseSSL); useSSL != "" {
		cfg.UseSSL = parseBool(useSSL, true)
	}

	if pathStyle := os.Getenv(EnvS3PathStyle); pathStyle != "" {
		cfg.PathStyle = parseBool(pathStyle, false)
	}

	if connectTimeout := os.Getenv(EnvS3ConnectTimeout); connectTimeout != "" {
		if d, err := time.ParseDuration(connectTimeout); err == nil {
			cfg.ConnectTimeout = d
		}
	}

	if requestTimeout := os.Getenv(EnvS3RequestTimeout); requestTimeout != "" {
		if d, err := time.ParseDuration(requestTimeout); err == nil {
			cfg.RequestTimeout = d
		}
	}

	if maxRetries := os.Getenv(EnvS3MaxRetries); maxRetries != "" {
		if n, err := strconv.Atoi(maxRetries); err == nil {
			cfg.MaxRetries = n
		}
	}

	if debug := os.Getenv(EnvS3Debug); debug != "" {
		cfg.Debug = parseBool(debug, false)
	}

	return cfg
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
