package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/clients/s3client"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/common/pkg/serialize"

	"github.com/vitistack/common/pkg/settings/dotenv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	trueString  = "true"
	falseString = "false"
)

func main() {
	dotenv.LoadDotEnv()

	logJsonEnabled := os.Getenv("LOG_JSON_ENABLED") // example usage of an env var loaded from .env
	vlog.Infof("LOG_JSON_ENABLED: %s", logJsonEnabled)
	logColorizeEnabled := os.Getenv("LOG_COLORIZE_ENABLED")
	vlog.Infof("LOG_COLORIZE_ENABLED: %s", logColorizeEnabled)
	logAddCaller := os.Getenv("LOG_ADD_CALLER")
	vlog.Infof("LOG_ADD_CALLER: %s", logAddCaller)
	logLevel := os.Getenv("LOG_LEVEL")
	vlog.Infof("LOG_LEVEL: %s", logLevel)
	logUnescapeMultiline := os.Getenv("LOG_UNESCAPE_MULTILINE")
	vlog.Infof("LOG_UNESCAPE_MULTILINE: %s", logUnescapeMultiline)
	logDisableStacktrace := os.Getenv("LOG_DISABLE_STACKTRACE")
	vlog.Infof("LOG_DISABLE_STACKTRACE: %s", logDisableStacktrace)

	// Initialize the logger
	err := vlog.Setup(vlog.Options{
		Level:             logLevel,                         // debug|info|warn|error|dpanic|panic|fatal
		ColorizeLine:      logColorizeEnabled == trueString, // whole-line color
		JSON:              logJsonEnabled == trueString,     // console output (supports ANSI colors)
		AddCaller:         logAddCaller == trueString,
		DisableStacktrace: logDisableStacktrace == trueString,
		UnescapeMultiline: logUnescapeMultiline == trueString, // unescape multiline messages (makes them more readable)
	})
	if err != nil {
		panic(err)
	}
	defer func() { _ = vlog.Sync() }()

	vlog.Info("This is an info message")
	vlog.Debug("This is a debug message")
	vlog.Warn("This is a warning message")
	vlog.Error("This is an error message")

	test := 42
	vlog.Debug("Log line ", "with extra parameters ", test)

	// Initialize Kubernetes client
	k8sclient.Init()
	vlog.Info("Kubernetes client initialized successfully")
	pods, err := k8sclient.Kubernetes.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		vlog.Error("Failed to list pods", err)
		return
	}
	for i := range pods.Items {
		pod := pods.Items[i]
		vlog.Debug("Pod:", pod.Name, "Pod labels:", serialize.Pretty(pod.Labels))
	}
	vlog.Info("Number of pods in default namespace:", len(pods.Items))

	// create test struct and log it
	testStruct := TestStruct{
		Field1: "value1",
		Field2: 123,
		Field3: TestStruct2{
			SubField1: "subvalue1",
			SubField2: 456,
		},
	}
	vlog.Info("Test struct:", serialize.Pretty(testStruct))

	// Demonstrate auto-formatting of JSON structures (new feature)
	vlog.Info("Auto-formatted struct", "data", testStruct)

	// Demonstrate auto-formatted map
	configMap := map[string]any{
		"database": map[string]string{
			"host": "localhost",
			"port": "5432",
		},
		"features": []string{"auth", "api", "cache"},
		"debug":    true,
	}
	vlog.Info("Auto-formatted map", "config", configMap)

	// Demonstrate JSON string auto-formatting
	jsonResponse := `{"status":"success","data":{"id":1,"name":"test"},"metadata":{"version":"1.0"}}`
	vlog.Info("Auto-formatted JSON string", "response", jsonResponse)

	// Demonstrate YAML formatting
	vlog.Info("YAML formatted struct", "yaml", serialize.YAML(testStruct))
	vlog.Info("YAML formatted map", "yaml", serialize.PrettyYAML(configMap))

	// S3 Client Example
	demonstrateS3Client()
}

// demonstrateS3Client shows S3 client usage with mock or real implementation.
func demonstrateS3Client() {
	// By default, use mock client (S3_MOCK=true or unset)
	// Set S3_MOCK=false to use real S3 client
	ctx := context.TODO()
	var s3 s3client.S3Client

	useMock := os.Getenv(s3client.EnvS3Mock)
	bucket := s3client.GetBucketFromEnv() // Read from S3_BUCKET env var

	if useMock == "" || useMock == trueString {
		// Use mock client (default)
		vlog.Info("Using mock S3 client (set S3_MOCK=false to use real client)")
		s3 = s3client.NewMockS3ClientFromEnv()
		bucket = "example-bucket" // Mock client bucket
	} else {
		// Use real S3 client
		vlog.Info("Using real S3 client")
		var err error
		realClient, err := s3client.NewGenericS3ClientFromEnv()
		if err != nil {
			vlog.Error("Failed to create S3 client from env", err)
			return
		}
		s3 = realClient

		// Log the configuration for debugging
		cfg := realClient.GetConfig()
		vlog.Debug("S3 Client Configuration",
			"endpoint", cfg.Endpoint,
			"region", cfg.Region,
			"useSSL", cfg.UseSSL,
			"insecureSkipVerify", cfg.InsecureSkipVerify,
			"pathStyle", cfg.PathStyle,
			"connectTimeout", cfg.ConnectTimeout,
			"requestTimeout", cfg.RequestTimeout,
		)
	}
	defer func() { _ = s3.Close() }()

	vlog.Info("Using bucket", "bucket", bucket)

	// For mock client, we need to create the bucket first and add sample files
	// Real S3 bucket should already exist
	if useMock == "" || useMock == trueString {
		if err := s3.CreateBucket(ctx, bucket); err != nil {
			vlog.Warn("Bucket may already exist", "error", err)
		}

		// Add sample files to the mock bucket
		sampleFiles := []struct {
			key         string
			content     string
			contentType string
		}{
			{"documents/readme.txt", "This is a sample readme file.", "text/plain"},
			{"documents/notes.md", "# Notes\n\nSome markdown notes.", "text/markdown"},
			{"config/settings.json", `{"environment": "dev", "debug": true}`, "application/json"},
			{"config/app.yaml", "name: example-app\nversion: 1.0.0", "application/x-yaml"},
			{"backups/data-2026-01-28.tar.gz", "fake-compressed-data", "application/gzip"},
		}

		for _, file := range sampleFiles {
			content := []byte(file.content)
			_, err := s3.PutObject(ctx, bucket, file.key, bytes.NewReader(content), int64(len(content)),
				s3client.WithContentType(file.contentType),
			)
			if err != nil {
				vlog.Error(fmt.Sprintf("Failed to add sample file %s", file.key), err)
			} else {
				vlog.Debug("Added sample file", "key", file.key)
			}
		}
	}

	// Check if bucket exists
	exists, err := s3.BucketExists(ctx, bucket)
	if err != nil {
		vlog.Error("Failed to check bucket exists", err)
		return
	}
	vlog.Info("Bucket exists", "bucket", bucket, "exists", exists)

	objList, err := s3.ListObjects(ctx, bucket)
	if err != nil {
		vlog.Error("Failed to list S3 objects", err)
		return
	}
	vlog.Info("Objects in bucket", "count", len(objList.Objects))
	for _, obj := range objList.Objects {
		vlog.Info("S3 Object:", "Key", obj.Key, "Size", obj.Size, "LastModified", obj.LastModified)
	}
}

type TestStruct struct {
	Field1 string      `json:"field1,omitempty" yaml:"field1,omitempty"`
	Field2 int         `json:"field2,omitempty" yaml:"field2,omitempty"`
	Field3 TestStruct2 `json:"field3,omitempty" yaml:"field3,omitempty"`
}

type TestStruct2 struct {
	SubField1 string `json:"subField1,omitempty" yaml:"subField1,omitempty"`
	SubField2 int    `json:"subField2,omitempty" yaml:"subField2,omitempty"`
}
