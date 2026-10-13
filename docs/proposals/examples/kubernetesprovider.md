# `KubernetesProvider`

| | |
|---|---|
| **API** | `vitistack.io/v1alpha3` |
| **Scope** | Cluster |
| **Spec written by** | The provider operator (talos-operator); never users |
| **Status written by** | The provider operator |
| **Referenced by** | `KubernetesClusterClass.spec.kubernetesProviderRef` |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §3.1 |

Metadata registration of what a cluster provider offers in this datacenter; sent upstream for visualisation. Carries no desired state for clusters.

The fields below are the subset of today's type (`common/pkg/v1alpha1/kubernetes_provider.go`) that fits a registration. Today's type also has `cluster`, `nodePools`, `network`, `addons` and similar config blocks, which §3.1 moves to classes and provider parameters.

## Full specification

```yaml
apiVersion: vitistack.io/v1alpha3
kind: KubernetesProvider
metadata:
  name: talos-operator
spec:
  # Required. talos | aks. Must match the controllerName of classes that reference it.
  providerType: talos
  # Required. Human-readable name.
  displayName: Talos (Oslo)
  # Required. Version of the provider operator or platform.
  version: v1.13.10
  # Required. Region this provider serves.
  region: oslo
  # Optional. Zones available for placement.
  zones: [oslo-a, oslo-b]
  # Optional. Machine backend this provider provisions nodes on.
  machineProviderRef:
    name: kubevirt-prod-1
  # Optional. Free-form provider metadata shown upstream.
  config:
    supportedKubernetesVersions: v1.31,v1.32,v1.33
status:
  # Pending | Ready | Failed | Offline.
  phase: Ready
  message: ""
  # Nodes managed by this provider.
  nodeCount: 42
  conditions:
    - { type: Ready, status: "True", reason: OperatorRunning }
```

## Minimal

```yaml
apiVersion: vitistack.io/v1alpha3
kind: KubernetesProvider
metadata:
  name: talos-operator
spec:
  providerType: talos
  displayName: Talos
  version: v1.13.10
  region: oslo
```
