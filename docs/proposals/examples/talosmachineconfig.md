# `TalosMachineConfig`

| | |
|---|---|
| **API** | `talos.vitistack.io/v1alpha3` |
| **Scope** | Namespace |
| **Spec written by** | Owner of the Machine (NodePool controller or cluster provider); never users |
| **Status written by** | OS provider (talos-operator) |
| **Owner** | Its `Machine` |
| **Owns** | `<machine>-bootstrap` Secret |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §6.2–§6.4 |

The Talos configuration of one Machine. Created from a `TalosMachineConfigTemplate`; updated in place for `InPlace` rollouts. The machine provider reads only its status.

## Full specification

```yaml
apiVersion: talos.vitistack.io/v1alpha3
kind: TalosMachineConfig
metadata:
  # Same name as the Machine.
  name: prod-a-gpu-7f9c-x2k4
  namespace: tenant-a
  labels:
    vitistack.io/cluster: prod-a
    vitistack.io/nodepool: prod-a-gpu
    vitistack.io/role: worker
  ownerReferences:
    - { apiVersion: vitistack.io/v1alpha3, kind: Machine, name: prod-a-gpu-7f9c-x2k4, uid: 4c1d…, controller: true }
spec:
  # Generic contract fields (§6.2), written by the owner of the Machine.
  clusterName: prod-a
  machineName: prod-a-gpu-7f9c-x2k4
  # controlplane | worker
  role: worker
  # Kubernetes version; pool pin, else the cluster's. Changing it = in-place Kubernetes upgrade.
  version: v1.33.2
  # Talos version; pool pin, else the cluster's. Changing it = in-place Talos upgrade.
  osVersion: v1.13.10
  # Resolved template this object was created from (§6.1).
  templateRef:
    { group: talos.vitistack.io, kind: TalosMachineConfigTemplate, name: gpu, namespace: tenant-a }
  # Cluster secrets bundle (§7); read only by the OS provider.
  secretsRef: { name: prod-a-talos-secrets }
  # Copied from the template's spec.template.spec (see talosmachineconfigtemplate.md).
  image:
    factory: factory.talos.dev
    platform: nocloud
    addons: [siderolabs/nonfree-kmod-nvidia-production, siderolabs/nvidia-container-toolkit-production]
    extraKernelArgs: [net.ifnames=1]
  install: { diskSelector: { size: ">= 100GB", type: ssd }, wipe: false }
  kernel: { modules: [{ name: nvidia }, { name: nvidia_uvm }] }
  kubelet: { nodeLabels: { nvidia.com/gpu.present: "true" } }
  registries: { mirrors: { docker.io: { endpoints: ["https://registry.internal/v2/docker.io"] } } }
  time: { servers: [ntp1.internal] }
  configPatches: []
status:
  observedGeneration: 1
  # Generic contract (§6.2) -------------------------------------------------
  # Bootstrap step finished; the machine may boot.
  ready: true
  # Secret with pre-boot data, read by the machine provider; empty for postBoot.
  dataSecretName: prod-a-gpu-7f9c-x2k4-bootstrap
  # nocloud | configDrive | ignition | kernelArgsURL
  dataFormat: nocloud
  # Boot or installer image for the desired OS version.
  image: factory.talos.dev/nocloud-installer/9a3c…:v1.13.10
  # How configuration reaches this node (§6.4).
  lifecycle: { initial: preBoot, updates: inPlace }
  # Hash of the desired config, and of what the node reports.
  renderedHash: sha256:3f1c…
  appliedHash: sha256:3f1c…
  # Talos-specific ----------------------------------------------------------
  # Image Factory schematic for addons (including required extensions) and extraKernelArgs.
  schematicID: 9a3c…
  # Extensions the node reports as installed.
  addons: [siderolabs/iscsi-tools, siderolabs/nonfree-kmod-nvidia-production, siderolabs/nvidia-container-toolkit-production, siderolabs/qemu-guest-agent]
  conditions:
    # Accepted=False reason: UnsupportedVersion (§10.5).
    - { type: Accepted,            status: "True",  reason: Accepted }
    - { type: Rendered,            status: "True",  reason: Rendered }
    - { type: Applied,             status: "True",  reason: HashesMatch }
    - { type: RequiresReplacement, status: "False", reason: InPlaceSupported }
```

## Minimal

Users don't write this object. This is the smallest object an owner writes: the contract fields only, with no template config.

```yaml
apiVersion: talos.vitistack.io/v1alpha3
kind: TalosMachineConfig
metadata:
  name: dev-a-cp-1
  namespace: tenant-a
spec:
  clusterName: dev-a
  machineName: dev-a-cp-1
  role: controlplane
  version: v1.33.2
  osVersion: v1.13.10
  templateRef: { group: talos.vitistack.io, kind: TalosMachineConfigTemplate, name: standard-cp, namespace: platform }
  secretsRef: { name: dev-a-talos-secrets }
```
