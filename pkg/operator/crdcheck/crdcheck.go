package crdcheck

import (
	"context"
	"fmt"
	"strings"

	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
)

// Ref describes a CRD-backed resource that must be available on the cluster.
// Provide the API group, version, and plural resource name, e.g.:
//
//	Ref{Group: "batch", Version: "v1", Resource: "jobs"}
//	Ref{Group: "example.com", Version: "v1alpha1", Resource: "widgets"}
type Ref struct {
	Group    string
	Version  string
	Resource string // plural (as used in kubectl, e.g., "widgets")
}

func (r Ref) String() string { return fmt.Sprintf("%s/%s %s", r.Group, r.Version, r.Resource) }

// EnsureInstalled checks that all provided CRD resources exist via the Discovery API.
// It returns an error listing any that are missing.
func EnsureInstalled(ctx context.Context, dc discovery.DiscoveryInterface, crds []Ref) error {
	if len(crds) == 0 {
		return nil
	}

	var missing []string
	for _, ref := range crds {
		gv := fmt.Sprintf("%s/%s", ref.Group, ref.Version)
		// Query the server for resources served under this group/version.
		list, err := dc.ServerResourcesForGroupVersion(gv)
		if err != nil {
			// NotFound means the entire GV isn't served; consider missing.
			if errors.IsNotFound(err) {
				vlog.Warnf("Required API group/version not found: %s (for resource %s)", gv, ref.Resource)
				missing = append(missing, ref.String())
				continue
			}
			// Any other error (e.g., permissions, connectivity) is treated as fatal for certainty.
			return fmt.Errorf("failed to discover resources for %s: %w", gv, err)
		}

		// Scan for the specific plural resource name.
		found := false
		for i := range list.APIResources {
			apiRes := &list.APIResources[i]
			if apiRes.Name == ref.Resource { // plural match
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, ref.String())
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required CRDs/resources: %s", strings.Join(missing, ", "))
	}
	return nil
}

// MustEnsureInstalled checks the provided CRDs using the global DiscoveryClient
// and panics if any are missing. Suitable to call during operator startup.
func MustEnsureInstalled(ctx context.Context, crds ...Ref) {
	if k8sclient.DiscoveryClient == nil {
		vlog.Error("Discovery client is not initialized; call k8sclient.Init() first")
		panic("k8s discovery client not initialized")
	}
	if err := EnsureInstalled(ctx, k8sclient.DiscoveryClient, crds); err != nil {
		vlog.Error("Required CRDs are not installed:", err)
		panic(err)
	}
	// Log success with a concise list
	if len(crds) > 0 {
		items := make([]string, 0, len(crds))
		for _, r := range crds {
			items = append(items, r.String())
		}
		vlog.Infof("All required CRDs/resources are installed: %s", strings.Join(items, ", "))
	}
}
