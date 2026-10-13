# `BareMetalMachineParameters` (future)

| | |
|---|---|
| **API** | `baremetal.vitistack.io/v1alpha3` |
| **Scope** | Cluster (referenced from cluster-scoped MachineClasses) |
| **Spec written by** | Platform admin |
| **Status written by** | — (no status) |
| **Read by** | Bare-metal provider |
| **Proposal** | [physical-machines.md](../physical-machines.md) §4.1 |

Bare-metal settings behind a `MachineClass`: which hosts may be claimed and how they boot.

## Full specification

```yaml
apiVersion: baremetal.vitistack.io/v1alpha3
kind: BareMetalMachineParameters
metadata:
  name: gpu
spec:
  # Required. Which PhysicalHosts may be claimed, by label.
  hostSelector:
    matchLabels: { hardware.vitistack.io/gpu: a100 }
    matchExpressions:
      - { key: vitistack.io/zone, operator: In, values: [oslo-a, oslo-b] }
  # How the image is delivered: virtualMedia | pxe.
  bootMethod: virtualMedia
  # Bootstrap data formats this provider can deliver; checked against the OS config's status.dataFormat.
  supportedBootstrapFormats: [nocloud, configDrive, kernelArgsURL]
  # Default reclaim policy for claimed hosts: Wipe | Retain.
  reclaimPolicy: Wipe
  # Prefer the host of the Machine being replaced (keeps local storage and cabling): PreferPrevious | None.
  hostAffinity: PreferPrevious
  # Spread Machines of one pool or control plane across failure domains.
  topologySpread:
    - { topologyKey: vitistack.io/rack, maxSkew: 1 }
```

## Minimal

```yaml
apiVersion: baremetal.vitistack.io/v1alpha3
kind: BareMetalMachineParameters
metadata:
  name: general
spec:
  hostSelector:
    matchLabels: { vitistack.io/zone: oslo-a }
```
