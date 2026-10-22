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

package v1alpha2

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

// TestAddToSchemeRegistersMetaTypes guards against a missing
// metav1.AddToGroupVersion call in addKnownTypes. Without it, client-go cannot
// build List/Watch requests and the controller cache fails at runtime with:
//
//	no kind "ListOptions" is registered for version "vitistack.io/v1alpha2"
func TestAddToSchemeRegistersMetaTypes(t *testing.T) {
	s := runtime.NewScheme()
	if err := AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}

	for _, kind := range []string{"ListOptions", "GetOptions", "DeleteOptions", "WatchEvent"} {
		gvk := GroupVersion.WithKind(kind)
		if !s.Recognizes(gvk) {
			t.Errorf("scheme does not recognize %s; is metav1.AddToGroupVersion missing from addKnownTypes?", gvk)
		}
	}
}
