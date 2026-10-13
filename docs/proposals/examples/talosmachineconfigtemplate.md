# `TalosMachineConfigTemplate`

| | |
|---|---|
| **API** | `talos.vitistack.io/v1alpha3` |
| **Scope** | Namespace |
| **Spec written by** | Platform admin (class defaults) or user (cluster or pool overrides) |
| **Status written by** | — (no status) |
| **Read by** | Cluster provider and NodePool controller, when creating `TalosMachineConfig` objects |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §6.1, §6.2 |

Talos machine configuration for a control plane or pool. Holds no version: the Talos version comes from `osVersion` on the cluster or pool (§10.1). A new `generation` rolls out to the Machines using it.

## Full specification

```yaml
apiVersion: talos.vitistack.io/v1alpha3
kind: TalosMachineConfigTemplate
metadata:
  name: gpu
  namespace: tenant-a
spec:
  template:
    # Copied as-is into each TalosMachineConfig.spec.
    spec:
      image:
        # Image Factory used to build the schematic.
        factory: factory.talos.dev
        # Talos platform; must match a dataFormat the machine provider can deliver (nocloud, metal, …).
        platform: nocloud
        # Talos system extensions baked into the image; cluster-required extensions are added.
        addons:
          - siderolabs/nonfree-kmod-nvidia-production
          - siderolabs/nvidia-container-toolkit-production
        # Extra kernel arguments baked into the image.
        extraKernelArgs: [net.ifnames=1]
      install:
        # Selects the install disk from Machine.status.hardware.disks.
        diskSelector: { size: ">= 100GB", type: ssd }
        # Wipe the install disk before installing. Default false.
        wipe: false
      kernel:
        # Kernel modules to load.
        modules:
          - { name: nvidia }
          - { name: nvidia_uvm }
          - { name: nvidia_drm }
          - { name: nvidia_modeset }
        # Kernel parameters.
        sysctls: { net.core.bpf_jit_harden: "1" }
      kubelet:
        # Extra kubelet flags.
        extraArgs: { max-pods: "110", rotate-server-certificates: "true" }
        # KubeletConfiguration fields merged into the generated config.
        extraConfig: { serializeImagePulls: false }
        # Labels the kubelet registers the Node with.
        nodeLabels: { nvidia.com/gpu.present: "true" }
        # Host paths mounted into the kubelet.
        extraMounts:
          - { source: /var/lib/nvidia, destination: /var/lib/nvidia, type: bind, options: [bind, rshared, rw] }
      registries:
        # Pull-through mirrors per registry.
        mirrors:
          docker.io: { endpoints: ["https://registry.internal/v2/docker.io"] }
      time:
        # NTP servers.
        servers: [ntp1.internal, ntp2.internal]
      # Escape hatch: raw Talos patches, applied last.
      configPatches:
        - |
          machine:
            features:
              hostDNS: { enabled: true, forwardKubeDNSToHost: true }
```

## Minimal

Talos defaults for a VM delivered with nocloud data.

```yaml
apiVersion: talos.vitistack.io/v1alpha3
kind: TalosMachineConfigTemplate
metadata:
  name: standard-worker
  namespace: platform
spec:
  template:
    spec:
      image:
        platform: nocloud
```
