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

// Package conversion implements the CRD conversion webhook for NetworkNamespace
// between v1alpha1 and v1alpha2.
package conversion

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	v1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	v1alpha2 "github.com/vitistack/common/pkg/v1alpha2"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const statusFailure = "Failure"

var log = logf.Log.WithName("conversion-webhook")

// Handler returns an http.Handler that serves the CRD conversion webhook
// for NetworkNamespace resources.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("reading request body: %w", err))
			return
		}

		review := &apiextensionsv1.ConversionReview{}
		if err := json.Unmarshal(body, review); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("decoding ConversionReview: %w", err))
			return
		}

		if review.Request == nil {
			writeError(w, http.StatusBadRequest, errors.New("ConversionReview.request is nil"))
			return
		}

		response := convert(review.Request)
		review.Response = response
		review.Response.UID = review.Request.UID
		review.Request = nil

		respBytes, err := json.Marshal(review)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("encoding response: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			log.Error(err, "failed to write response")
		}
	})
}

func convert(req *apiextensionsv1.ConversionRequest) *apiextensionsv1.ConversionResponse {
	resp := &apiextensionsv1.ConversionResponse{
		ConvertedObjects: make([]runtime.RawExtension, 0, len(req.Objects)),
	}

	for i, raw := range req.Objects {
		obj := &unstructured.Unstructured{}
		if err := json.Unmarshal(raw.Raw, &obj.Object); err != nil {
			resp.Result = metav1.Status{
				Status:  statusFailure,
				Message: fmt.Sprintf("decoding object %d: %v", i, err),
			}
			return resp
		}

		srcVersion := obj.GetAPIVersion()
		dstVersion := req.DesiredAPIVersion

		if srcVersion == dstVersion {
			// No conversion needed, pass through
			resp.ConvertedObjects = append(resp.ConvertedObjects, raw)
			continue
		}

		converted, err := convertObject(obj, srcVersion, dstVersion)
		if err != nil {
			resp.Result = metav1.Status{
				Status:  statusFailure,
				Message: fmt.Sprintf("converting object %d from %s to %s: %v", i, srcVersion, dstVersion, err),
			}
			return resp
		}

		rawConverted, err := json.Marshal(converted)
		if err != nil {
			resp.Result = metav1.Status{
				Status:  statusFailure,
				Message: fmt.Sprintf("encoding converted object %d: %v", i, err),
			}
			return resp
		}
		resp.ConvertedObjects = append(resp.ConvertedObjects, runtime.RawExtension{Raw: rawConverted})
	}

	resp.Result = metav1.Status{Status: metav1.StatusSuccess}
	return resp
}

func convertObject(obj *unstructured.Unstructured, srcVersion, dstVersion string) (*unstructured.Unstructured, error) {
	kind := obj.GetKind()
	if kind != "NetworkNamespace" {
		return nil, fmt.Errorf("unsupported kind %q for conversion", kind)
	}

	switch {
	case srcVersion == "vitistack.io/v1alpha1" && dstVersion == "vitistack.io/v1alpha2":
		return convertV1alpha1ToV1alpha2(obj)
	case srcVersion == "vitistack.io/v1alpha2" && dstVersion == "vitistack.io/v1alpha1":
		return convertV1alpha2ToV1alpha1(obj)
	default:
		return nil, fmt.Errorf("unsupported conversion from %s to %s", srcVersion, dstVersion)
	}
}

func convertV1alpha1ToV1alpha2(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	// Deserialize to typed v1alpha1
	raw, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, fmt.Errorf("marshalling unstructured: %w", err)
	}

	src := &v1alpha1.NetworkNamespace{}
	if err := json.Unmarshal(raw, src); err != nil {
		return nil, fmt.Errorf("unmarshalling v1alpha1 NetworkNamespace: %w", err)
	}

	// Convert to typed v1alpha2
	dst := v1alpha2.ConvertNetworkNamespaceFromV1alpha1(src)

	// Serialize back to unstructured
	dstRaw, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshalling v1alpha2: %w", err)
	}

	result := &unstructured.Unstructured{}
	if err := json.Unmarshal(dstRaw, &result.Object); err != nil {
		return nil, fmt.Errorf("unmarshalling to unstructured: %w", err)
	}

	result.SetAPIVersion("vitistack.io/v1alpha2")
	return result, nil
}

func convertV1alpha2ToV1alpha1(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	raw, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, fmt.Errorf("marshalling unstructured: %w", err)
	}

	src := &v1alpha2.NetworkNamespace{}
	if err := json.Unmarshal(raw, src); err != nil {
		return nil, fmt.Errorf("unmarshalling v1alpha2 NetworkNamespace: %w", err)
	}

	dst := v1alpha2.ConvertNetworkNamespaceToV1alpha1(src)

	dstRaw, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshalling v1alpha1: %w", err)
	}

	result := &unstructured.Unstructured{}
	if err := json.Unmarshal(dstRaw, &result.Object); err != nil {
		return nil, fmt.Errorf("unmarshalling to unstructured: %w", err)
	}

	result.SetAPIVersion("vitistack.io/v1alpha1")
	return result, nil
}

func writeError(w http.ResponseWriter, code int, err error) {
	log.Error(err, "conversion webhook error", "code", code)
	http.Error(w, err.Error(), code)
}
