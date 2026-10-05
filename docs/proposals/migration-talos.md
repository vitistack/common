# Migration: from today's Talos implementation to the Kubernetes-native API

| | |
|---|---|
| **Status** | Draft / for discussion |
| **Target design** | [kubernetes-native-cluster-machine-api.md](kubernetes-native-cluster-machine-api.md) |
| **Affects** | `common`, talos-operator, kubevirt-operator |
| **Date** | 2026-10-05 |

## 1. Current situation

### 1.1 talos-operator

- **One reconcile loop** for `KubernetesCluster` that owns `Machine` objects and requeues every cluster every 5 s, also when nothing has changed.
- Generates the Talos config and **pushes** it to each node after boot, while the node is in maintenance mode. This needs insecure connections, waiting for reboots (inside the reconcile) and probing for maintenance mode.
- **Progress is stored as string flags in a per-cluster Secret**, next to credentials, generated configs and the kubeconfig. The cluster status partly mirrors it.
- **Upgrades and resets are triggered by annotations** on `KubernetesCluster`.
- **Tenant configuration comes from one global ConfigMap**, the same for all clusters and not watched. Changes reach new nodes only.
- Cluster-wide settings (Talos version, endpoint mode, extensions) come from operator environment variables.

### 1.2 kubevirt-operator

Contains implicit coupling to Talos:

- a no-op `#cloud-config` so Talos stays in maintenance mode
- treats a missing key in talos-operator's cluster Secret as "not ready yet"
- removes the boot ISO when the `vitistack.io/os-installed` annotation appears
- writes `Machine.status` by **replacing the whole status**, so no other operator can safely report on a Machine

### 1.3 The gap

```
KubernetesCluster  ──►  ???  ──►  Machine
 "which cluster"        "which OS config"   "which VM / host"
```

Every operating system needs per-machine configuration that depends on the cluster (endpoint, PKI, join credentials) and on the machine (role, hardware). Today no resource owns it; talos-operator fills the gap with API calls and kubevirt-operator with Talos-specific shortcuts.

## 2. Mapping: today → target

| Today | Target (main proposal) |
|---|---|
| `spec.data.provider: talos` | `spec.className` → `KubernetesClusterClass` with `controllerName` |
| Operator env vars + global tenant ConfigMap | `TalosClusterConfig` via class `parametersRef`, overridable per cluster |
| `spec.topology.workers.nodePools` | `spec.workers.pools` generating `NodePool` objects, or standalone NodePools |
| Secret flags (`bootstrapped`, `worker_applied`, ...) | `KubernetesCluster.status.conditions` |
| `configured_nodes` set in the Secret | `TalosMachineConfig.status` per Machine (`appliedHash`, conditions) |
| Upgrade annotations | `spec.version` (Kubernetes); Talos version in `TalosClusterConfig` / machine config templates |
| Push config in maintenance mode | Pre-boot nocloud data from `TalosMachineConfig`; push only as fallback |
| `vitistack.io/os-installed` annotation | `OSConfigured` condition on `Machine` |
| Scale-down code path before Machine deletion | Cluster provider's finalizer on each `Machine` |
| One Secret with everything | Standard Secrets: `<cluster>-kubeconfig`, `<cluster>-ca`, `<cluster>-talos-secrets`, `<machine>-bootstrap` |

## 3. Steps

Each step is releasable on its own and keeps existing clusters working.

| Step | Change | Existing clusters | Result |
|---|---|---|---|
| 1 | **kubevirt-operator writes status with server-side apply** and owns only its fields and conditions | Unaffected | Other operators can report on Machines |
| 2 | **Add generic types to `common`**: `KubernetesClusterClass`, `MachineClass.controllerName`, `NodePool`, `Machine.spec.bootstrap.configRef`, Machine conditions, OS config status contract | Unaffected: Machines without `bootstrap` behave as today | Contract exists |
| 3 | **kubevirt-operator honours the bootstrap contract**: wait for `BootstrapReady`, boot `status.image` with `status.dataSecretName`, remove the ISO on `OSConfigured`. Delete VM only when its finalizer is the last one | Unaffected (no `bootstrap` set) | Machine provider is OS-agnostic for new Machines |
| 4 | **talos-operator gains an OS-provider controller** for `TalosMachineConfig` / `TalosMachineConfigTemplate`: render per-machine config, pre-boot nocloud delivery on VMs, push only as fallback | New Machines only | No maintenance-mode push or reboot waits for new nodes |
| 5 | **Cluster loop moves state to status**: one-time import of Secret flags and `configured_nodes` into `KubernetesCluster.status` and `TalosMachineConfig` objects for existing Machines; cleanup finalizer on Machines; standard Secret names | Migrated automatically; old flags kept read-only for one release for rollback | Secret holds only credentials |
| 6 | **`TalosClusterConfig` and classes**: create `talos-standard` from current env vars and the tenant ConfigMap; set `className` on existing clusters | Migrated automatically | Per-cluster configuration, watched |
| 7 | **Topology**: generic NodePool and topology controllers; convert `spec.topology.workers.nodePools` to `spec.workers.pools` (new API version with conversion) | Converted; generated NodePools adopt existing worker Machines by label | `/scale`, per-pool rollout |
| 8 | **Upgrades via spec**: `spec.version` and Talos version in parameters drive upgrades; annotation path deprecated, then removed | Annotations honoured during deprecation | No commands in annotations |
| 9 | **Remove legacy**: Talos shortcuts in kubevirt-operator, the global ConfigMap path, Secret flags, annotation handling | - | Target design reached |
| 10 | **Split the OS provider** into its own operator when a second OS is added | - | Same contract, separate deployment |

## 4. Risks

| Risk | Mitigation |
|---|---|
| Losing progress state during import (step 5) | Import is idempotent; old flags stay readable for one release; import result recorded as a `Migrated` condition |
| Generated NodePools adopting the wrong Machines (step 7) | Adopt only Machines with matching cluster and pool labels and an owner reference to the cluster; dry-run report first |
| New pre-boot path fails on some VMs (step 4) | Push fallback stays until step 9; per-class opt-in |
| API version conversion (step 7) | Conversion webhook with round-trip tests before the new version becomes the storage version |

## 5. Open questions

| # | Question |
|---|---|
| 1 | New API version (`v1alpha2`) with conversion, or new resource names alongside the old ones? |
| 2 | How long do deprecated paths (annotations, Secret flags, push fallback) stay? |
| 3 | Order of steps 6 and 7: classes before topology, or together? |
