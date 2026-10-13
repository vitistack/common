# `NodePool`

| | |
|---|---|
| **API** | `vitistack.io/v1alpha3` |
| **Scope** | Namespace |
| **Spec written by** | User (standalone) or topology controller (generated); `replicas` via `/scale` by an autoscaler or user |
| **Status written by** | NodePool controller |
| **Owns** | Worker Machines |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §4.3, §4.4 |

A set of identical worker Machines, like a Deployment for Pods. Has a `/scale` subresource.

## Full specification

```yaml
apiVersion: vitistack.io/v1alpha3
kind: NodePool
metadata:
  name: batch-team-x
  namespace: tenant-a
spec:
  # Required. KubernetesCluster in the same namespace.
  clusterName: prod-a
  # Stops the rollout of this pool; Machines already updated stay. Default false.
  paused: false
  # Desired number of Machines; changed through /scale by an autoscaler or user.
  replicas: 2
  # Optional. Bounds enforced on /scale.
  autoscaling:
    minReplicas: 1
    maxReplicas: 5
  # Required. Must match template.metadata.labels.
  selector:
    matchLabels:
      vitistack.io/nodepool: batch-team-x
  # Required. MachineTemplateSpec: what Machines are created from.
  template:
    metadata:
      labels:
        vitistack.io/nodepool: batch-team-x
    spec:
      # Required. MachineClass for the Machines.
      className: medium
      # Optional pins; unset follows the cluster. Never newer than the control plane (§10.5).
      version: v1.33.2
      osVersion: v1.13.10
      # Optional. Overrides the cluster and class worker template; same namespace.
      machineConfig:
        templateRef:
          group: talos.vitistack.io
          kind: TalosMachineConfigTemplate
          name: batch
      # Optional. Labels and taints added to the Nodes.
      labels:
        workload: batch
      taints:
        - { key: batch, value: "true", effect: NoSchedule }
  strategy:
    # RollingUpdate (replace Machines) | InPlace (requires OS provider updates: inPlace, §6.4).
    type: RollingUpdate
    rollingUpdate:
      # Extra Machines allowed during a rollout.
      maxSurge: 1
      # Machines allowed to be unavailable during a rollout.
      maxUnavailable: 0
status:
  observedGeneration: 3
  replicas: 2
  readyReplicas: 2
  # Machines matching the current template hash.
  updatedReplicas: 2
  # Label selector string for the /scale subresource.
  selector: vitistack.io/nodepool=batch-team-x
  # Resolved template actually in use (§6.1).
  machineConfigTemplateRef:
    { group: talos.vitistack.io, kind: TalosMachineConfigTemplate, name: batch, namespace: tenant-a }
  conditions:
    # Accepted=False reasons: UnsupportedMachineConfigKind, NoMachineConfigTemplate, InPlaceNotSupported, VersionSkew.
    - { type: Accepted, status: "True", reason: Accepted }
    - { type: Ready,    status: "True", reason: AllReplicasReady }
```

## Minimal

Standalone pool; templates and versions follow the cluster and class.

```yaml
apiVersion: vitistack.io/v1alpha3
kind: NodePool
metadata:
  name: batch-team-x
  namespace: tenant-a
spec:
  clusterName: prod-a
  replicas: 2
  selector:
    matchLabels: { vitistack.io/nodepool: batch-team-x }
  template:
    metadata:
      labels: { vitistack.io/nodepool: batch-team-x }
    spec:
      className: medium
```
