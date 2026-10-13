# `TalosClusterConfig`

| | |
|---|---|
| **API** | `talos.vitistack.io/v1alpha3` |
| **Scope** | Namespace |
| **Spec written by** | Platform admin (class default) or user (per-cluster override) |
| **Status written by** | — (no status) |
| **Read by** | Cluster provider, OS provider (watched) |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §4.1; [migration-talos.md](../migration-talos.md) §2 |

Talos cluster-wide parameters. Replaces talos-operator's environment variables and the global tenant ConfigMap.

## Full specification

```yaml
apiVersion: talos.vitistack.io/v1alpha3
kind: TalosClusterConfig
metadata:
  name: standard
  namespace: platform
spec:
  endpoint:
    # How the control-plane endpoint is provided: networkconfiguration | talosvip | custom | none.
    mode: networkconfiguration
    # talosvip only: interface for the Layer 2 VIP.
    vipLink: net0
    # custom only: endpoint addresses.
    customAddresses: []
    # Operator that reconciles the ControlPlaneVirtualSharedIP: nam | static-ip-operator.
    loadBalancerProvider: nam
  # System extensions every node must have; added to each machine's addons by the OS provider.
  requiredExtensions:
    - siderolabs/iscsi-tools
    - siderolabs/qemu-guest-agent
    - siderolabs/trident-iscsi-tools
    - siderolabs/util-linux-tools
    - siderolabs/nfs-utils
  # Use predictable network interface names on all nodes.
  predictableNetworkNames: true
  # Cluster-level add-ons installed into the workload cluster.
  addons:
    # CNI: flannel | cilium | none (installed separately).
    cni: cilium
    # Extra manifests applied by Talos at bootstrap (inline or by URL).
    manifests:
      - name: metrics-server
        url: https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.7.2/components.yaml
  # Kubernetes control-plane component settings.
  apiServer:
    extraArgs: { audit-log-maxage: "30" }
    certSANs: [prod-a.example.internal]
  controllerManager:
    extraArgs: {}
  scheduler:
    extraArgs: {}
  etcd:
    extraArgs: {}
  # Raw Talos patches applied to every machine's config before the template's configPatches.
  configPatches:
    - |
      cluster:
        discovery: { enabled: true }
  # Mark a cluster Failed if no Machine is running after this long; 0 disables.
  provisionFailureTimeout: 15m
```

## Minimal

Every field has a default; an empty spec gives the operator defaults.

```yaml
apiVersion: talos.vitistack.io/v1alpha3
kind: TalosClusterConfig
metadata:
  name: standard
  namespace: platform
spec: {}
```
