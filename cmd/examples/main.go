package main

import (
	"context"
	"os"

	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/common/pkg/serialize"

	"github.com/vitistack/common/pkg/settings/dotenv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Level:             logLevel,                     // debug|info|warn|error|dpanic|panic|fatal
		ColorizeLine:      logColorizeEnabled == "true", // whole-line color
		JSON:              logJsonEnabled == "true",     // console output (supports ANSI colors)
		AddCaller:         logAddCaller == "true",
		DisableStacktrace: logDisableStacktrace == "true",
		UnescapeMultiline: logUnescapeMultiline == "true", // unescape multiline messages (makes them more readable)
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
}

type TestStruct struct {
	Field1 string      `json:"field1,omitempty"`
	Field2 int         `json:"field2,omitempty"`
	Field3 TestStruct2 `json:"field3,omitempty"`
}

type TestStruct2 struct {
	SubField1 string `json:"subField1,omitempty"`
	SubField2 int    `json:"subField2,omitempty"`
}
