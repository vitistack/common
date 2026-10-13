# `Machine`

| | |
|---|---|
| **API** | `vitistack.io/v1alpha3` |
| **Scope** | Namespace |
| **Spec written by** | Owner: NodePool controller, cluster provider, or a user (Machines without a cluster) |
| **Status written by** | Machine provider, OS provider and cluster provider, each with its own field manager |
| **Owns** | Its OS config object (e.g. `TalosMachineConfig`) |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §5.2 |

One VM or host. Mostly immutable; replaced rather than edited.

| Status writer | Fields and conditions |
|---|---|
| Machine provider | `addresses`, `hardware`, `hostRef`, `InfrastructureReady`, `Ready` |
| OS provider | `versions`, `BootstrapReady`, `OSConfigured` |
| Cluster provider | `nodeRef`, `NodeHealthy` |

## Full specification

```yaml
apiVersion: vitistack.io/v1alpha3
kind: Machine
metadata:
  name: prod-a-gpu-7f9c-x2k4
  namespace: tenant-a
  labels:
    vitistack.io/cluster: prod-a
    vitistack.io/nodepool: prod-a-gpu
    # controlplane | worker
    vitistack.io/role: worker
  ownerReferences:
    # NodePool for workers, KubernetesCluster for the control plane.
    - { apiVersion: vitistack.io/v1alpha3, kind: NodePool, name: prod-a-gpu, uid: 0b6f…, controller: true }
  finalizers:
    # Machine provider: deletes the VM; removed only when it is the last finalizer.
    - kubevirt.vitistack.io/machine
    # Cluster provider: drain, etcd membership, VIP.
    - talos.vitistack.io/node-cleanup
spec:                                        # immutable except providerID
  # Optional. KubernetesCluster this Machine belongs to.
  clusterName: prod-a
  # Required. MachineClass.
  className: gpu-xlarge
  # Optional. OS config object; same namespace. Without it the machine provider boots without data.
  bootstrap:
    configRef:
      group: talos.vitistack.io
      kind: TalosMachineConfig
      name: prod-a-gpu-7f9c-x2k4
  # Set once by the machine provider.
  providerID: kubevirt://kubevirt-prod-1/tenant-a/prod-a-gpu-7f9c-x2k4
status:
  observedGeneration: 1
  # Machine provider.
  addresses:
    - { type: InternalIP, address: 10.0.0.31 }
    - { type: Hostname,   address: prod-a-gpu-7f9c-x2k4 }
  # Machine provider; for physical machines from host inspection.
  hardware:
    interfaces: [{ name: eth0, mac: "52:54:00:12:34:56" }]
    disks: [{ path: /dev/vda, serial: abc123, sizeBytes: 107374182400 }]
  # Bare-metal provider only (physical-machines.md).
  hostRef: null
  # OS provider: versions actually running.
  versions:
    kubernetes: v1.33.2
    os: v1.13.10
  # Cluster provider.
  nodeRef: { name: prod-a-gpu-7f9c-x2k4 }
  conditions:
    - { type: InfrastructureReady, status: "True", reason: Running }      # machine provider
    - { type: BootstrapReady,      status: "True", reason: DataReady }    # OS provider
    - { type: OSConfigured,        status: "True", reason: Applied }      # OS provider
    - { type: NodeHealthy,         status: "True", reason: NodeReady }    # cluster provider
    - { type: Ready,               status: "True", reason: Ready }        # machine provider, from the above
```

## Minimal

A plain machine without a cluster or OS config.

```yaml
apiVersion: vitistack.io/v1alpha3
kind: Machine
metadata:
  name: build-runner-1
  namespace: tenant-a
spec:
  className: medium
```
