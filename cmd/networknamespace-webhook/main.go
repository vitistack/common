/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package main runs a standalone CRD conversion webhook for NetworkNamespace.
//
// It converts between every served version through the v1alpha3 hub, using
// the Hub/Convertible implementations in pkg/v1alpha1, pkg/v1alpha2 and
// pkg/v1alpha3. It needs no Kubernetes API access: it only answers the API
// server's ConversionReview requests on /convert.
package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/conversion"

	"github.com/vitistack/common/pkg/loggers/vlog"
	v1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	v1alpha2 "github.com/vitistack/common/pkg/v1alpha2"
	v1alpha3 "github.com/vitistack/common/pkg/v1alpha3"
)

func main() {
	var (
		port        int
		certDir     string
		certName    string
		keyName     string
		enableHTTP2 bool
	)
	flag.IntVar(&port, "port", 9443, "Webhook server port")
	flag.StringVar(&certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Directory containing TLS certs")
	flag.StringVar(&certName, "cert-name", "tls.crt", "TLS certificate file name")
	flag.StringVar(&keyName, "key-name", "tls.key", "TLS key file name")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"Serve HTTP/2. Off by default because of the HTTP/2 stream cancellation and rapid reset CVEs.")
	flag.Parse()

	_ = vlog.Setup(vlog.Options{Level: "info", JSON: true})
	log.SetLogger(vlog.Logr())
	logger := log.Log.WithName("networknamespace-conversion-webhook")

	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))
	utilruntime.Must(v1alpha3.AddToScheme(scheme))

	// Fail at startup, not on live requests, if a served version lacks a
	// conversion to the hub.
	if ok, err := conversion.IsConvertible(scheme, &v1alpha3.NetworkNamespace{}); err != nil || !ok {
		logger.Error(err, "NetworkNamespace is not convertible: every version must be the hub or implement conversion.Convertible")
		os.Exit(1)
	}

	// The serving certificate is provisioned by cert-manager into a mounted
	// Secret. Say so explicitly if it is missing, rather than fail opaquely.
	for _, name := range []string{certName, keyName} {
		f := filepath.Join(certDir, name)
		if _, err := os.Stat(f); err != nil {
			msg := "unable to access TLS certificate file"
			if errors.Is(err, fs.ErrNotExist) {
				msg = "TLS certificate not found: the serving cert is provisioned by cert-manager — " +
					"verify cert-manager (and its cainjector) is installed and the Certificate has been issued into the mounted Secret"
			}
			logger.Error(err, msg, "file", f)
			os.Exit(1)
		}
	}

	var tlsOpts []func(*tls.Config)
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, func(c *tls.Config) { c.NextProtos = []string{"http/1.1"} })
	}

	// The server reloads the certificate when cert-manager rotates it.
	srv := webhook.NewServer(webhook.Options{
		Port:     port,
		CertDir:  certDir,
		CertName: certName,
		KeyName:  keyName,
		TLSOpts:  tlsOpts,
	})
	srv.Register("/convert", conversion.NewWebhookHandler(scheme, conversion.NewRegistry()))
	srv.Register("/healthz", healthz.CheckHandler{Checker: healthz.Ping})
	srv.Register("/readyz", healthz.CheckHandler{Checker: srv.StartedChecker()})

	logger.Info("starting NetworkNamespace conversion webhook", "port", port)
	if err := srv.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "webhook server failed")
		os.Exit(1)
	}
}
