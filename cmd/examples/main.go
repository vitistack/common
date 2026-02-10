package main

import (
	"bytes"
	"context"
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
	kubernetesConfigSet := os.Getenv("KUBECONFIG")
	if kubernetesConfigSet != "" {
		getKubernetesPodsAndLogWithVLog(kubernetesConfigSet)
	}

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

	s3Example()
}

func getKubernetesPodsAndLogWithVLog(kubernetesConfigSet string) {
	vlog.Info("Using KUBECONFIG from environment:", kubernetesConfigSet)

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
}

func s3Example() {

	var s3 s3client.S3Client
	var err error

	if os.Getenv("S3_USE_MOCK") == trueString {
		vlog.Info("Using Mock S3 Client for testing")
		s3 = s3client.NewMockS3Client()
	} else {
		vlog.Info("Using real S3 Client with endpoint:", os.Getenv("S3_ENDPOINT"))
		s3, err = s3client.NewS3Client(
			s3client.WithEndpoint(os.Getenv("S3_ENDPOINT")),
			s3client.WithAccessKey(os.Getenv("S3_ACCESS_KEY")),
			s3client.WithSecretKey(os.Getenv("S3_SECRET_KEY")),
			s3client.WithBucketName(os.Getenv("S3_BUCKET_NAME")),
			s3client.WithSecure(os.Getenv("S3_SECURE") == trueString),
			s3client.WithRegion(os.Getenv("S3_REGION")),
		)
		if err != nil {
			vlog.Error("Failed to create real S3 client", err)
			return
		}
	}

	err = s3.CreateBucket(context.Background())
	if err != nil {
		vlog.Error("Failed to create bucket", err)
		return
	}
	vlog.Info("Bucket created successfully")

	// Create test data
	testData := []byte("Hello from S3! This is test data.")
	err = s3.PutObject(context.Background(), "test.txt", bytes.NewReader(testData), int64(len(testData)))
	if err != nil {
		vlog.Error("Failed to put object", err)
		return
	}
	vlog.Info("Object uploaded successfully")

	data, err := s3.GetObject(context.Background(), "test.txt")
	if err != nil {
		vlog.Error("Failed to get object", err)
		return
	}
	vlog.Info("Object data:", string(data))

	list, err := s3.ListObject(context.Background(), s3client.ListObjectsOptions{Prefix: "test", Recursive: true})
	if err != nil {
		vlog.Error("Failed to list objects", err)
		return
	}
	vlog.Info("Objects with prefix 'test':", serialize.Pretty(list))

	err = s3.DeleteObject(context.Background(), "test.txt")
	if err != nil {
		vlog.Error("Failed to delete object", err)
		return
	}
	vlog.Info("Object deleted successfully")

	err = s3.DeleteBucket(context.Background())
	if err != nil {
		vlog.Error("Failed to delete bucket", err)
		return
	}
	vlog.Info("Bucket deleted successfully")
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
