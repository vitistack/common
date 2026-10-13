# `MachineOperation`

| | |
|---|---|
| **API** | `vitistack.io/v1alpha3` |
| **Scope** | Namespace |
| **Spec written by** | User |
| **Status written by** | The provider that handles `spec.type` (split not yet specified) |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §8; optional for the initial scope (§13 #7) |

A one-off action on a Machine, like a Job. Auditable and RBAC-controlled.

## Full specification

```yaml
apiVersion: vitistack.io/v1alpha3
kind: MachineOperation
metadata:
  name: reboot-x2k4
  namespace: tenant-a
spec:
  # Required. Machine in the same namespace. Immutable.
  machineName: prod-a-gpu-7f9c-x2k4
  # Required. Reboot | Reset | ReapplyConfig. Immutable.
  type: Reboot
  # Optional. Drain the Node first when the Machine belongs to a cluster. Default true.
  drain: true
  # Optional. Give up after this long; the operation then reports Failed.
  timeout: 15m
status:
  observedGeneration: 1
  startTime: "2026-10-13T09:00:00Z"
  completionTime: "2026-10-13T09:03:12Z"
  conditions:
    # Complete or Failed is True when the operation has finished.
    - { type: Complete, status: "True", reason: Rebooted }
```

## Minimal

```yaml
apiVersion: vitistack.io/v1alpha3
kind: MachineOperation
metadata:
  name: reboot-x2k4
  namespace: tenant-a
spec:
  machineName: prod-a-gpu-7f9c-x2k4
  type: Reboot
```
