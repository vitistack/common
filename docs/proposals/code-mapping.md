# Proposal vs. current code

| | |
|---|---|
| **Status** | Draft / for discussion |
| **Relates to** | [kubernetes-native-cluster-machine-api.md](kubernetes-native-cluster-machine-api.md), [migration-talos.md](migration-talos.md) |
| **Date** | 2026-10-05 |

Paths are `repo/path#line` in the vitistack workspace.

## 1. Mapping

| Area | Today | Proposal | Effort |
|---|---|---|---|
| Cluster API shape | `spec.data` + `spec.topology` (`common/pkg/v1alpha1/kubernetesCluster.go#L42-L163`) | `className`, `version`, `controlPlane`, `workers.pools`, `network` | High: new API version + conversion |
| Cluster provider selection | `spec.data.provider` string, checked in reconcile (`talos-operator/api/controllers/v1alpha1/kubernetescluster_controller.go`) | `KubernetesClusterClass.controllerName` | Medium |
| Provider registration | Operators create `KubernetesProvider` / `MachineProvider` at startup (`talos-operator/internal/services/kubernetesproviderservice/kubernetes_provider_service.go#L206`, `kubevirt-operator/internal/services/initializationservice/initialization_service.go#L154`) | Admin-defined classes referencing the registrations, which are kept (main proposal §3.1) | Low |
| Talos settings | Env vars + global tenant ConfigMap (`talos-operator/internal/kubernetescluster/talos/talos_manager.go#L315`) | `TalosClusterConfig` via `parametersRef` | Medium |
| Machine creation | `MachineManager` per cluster pass (`talos-operator/internal/machine/machine_manager.go#L298`, `#L438`) | Cluster provider (control plane) + generic NodePool controller (workers) | Medium |
| Machine changes | `applyMachine` patches spec (`machine_manager.go#L586`); kubevirt only creates VMs (`kubevirt-operator/controllers/v1alpha1/machine_controller.go#L323-L388`) | Immutable Machines; changes roll out | Low: already immutable in practice. Today's spec patches are silently not applied |
| Machine provider selection | `Machine.spec.provider`, empty = kubevirt (`machine_controller.go#L720-L815`) | Class `controllerName` + `machineProviderRef` | Medium |
| Machine size | `MachineClass` = CPU/memory catalogue (`common/pkg/v1alpha1/machineclass.go#L41-L66`) | `MachineClass` = implementation + size + `machineProviderRef` | Medium: existing size classes become one class per size and backend |
| Backend selection | Annotation `vitistack.io/kubevirt-config` written by kubevirt-operator (`machine_controller.go#L135-L155`) | Class parameters or status | Low |
| Machine status writes | Whole status replaced (`kubevirt-operator/internal/machine/status/status_manager.go#L82-L103`) | Server-side apply per writer, `metav1.Condition` | Low, but a prerequisite |
| OS config delivery | Push over Talos API with reboot waits (`talos-operator/internal/services/talosclientservice/talos_client_service.go#L152`); no-op cloud-config and Secret-key wait in kubevirt (`kubevirt-operator/internal/machine/vm/cloudinit.go#L36-L42`, `#L276-L282`) | `TalosMachineConfig` + `BootstrapReady` | Medium |
| Facts before boot | MAC generated and stored before VM creation (`kubevirt-operator/internal/machine/vm/network_config.go#L72`); static IP awaited | `Machine.status.hardware` before boot | Low |
| Boot image | `Machine.spec.os` + annotations (`machine_manager.go#L43`, `#L85`) | OS config `status.image` | Low |
| ISO cleanup | `vitistack.io/os-installed` annotation (`kubevirt-operator/internal/machine/vm/cdrom_manager.go#L45-L49`) | `OSConfigured` condition | Low |
| Scale-down / deletion | `handleScaleDown` in cluster loop; kubevirt deletes VM immediately | Finalizer per party; VM deleted last | Medium |
| Progress state | Secret flags + `configured_nodes` (`talos-operator/internal/services/talosstateservice/talos_state_service.go#L175`, `#L960`) | Conditions + per-node OS config status | High (§2.1) |
| Cluster status | Unstructured updates, custom condition type (`talos-operator/internal/kubernetescluster/status/status_manager.go#L381`, `common/pkg/v1alpha1/kubernetesCluster.go#L251`) | One server-side-apply patch, `metav1.Condition` | Medium |
| Upgrades | 18 annotations (`talos-operator/pkg/consts/consts.go#L204-L317`) | `spec.version` + status; `MachineOperation` | Medium, external impact |
| Polling | 5 s requeue in both operators (`kubernetescluster_controller.go#L54`, `machine_controller.go#L184-L190`) | Event-driven | High for kubevirt (§2.6) |

## 2. Hard to migrate

| # | Issue | Why | Options |
|---|---|---|---|
| 1 | **Applied config on running Talos nodes** | Config was pushed at apply time with the tenant overrides of that moment; no record per node, so `appliedHash` is unknown | Read machine config from each node via Talos API and hash (normalised) · adopt current rendered config as applied · roll every node once |
| 2 | **`KubernetesCluster` API change** | All fields move; consumers outside this workspace (portal, API, automation, upgrade annotations) can't be verified | Conversion webhook; inventory consumers first |
| 3 | **`MachineClass` meaning** | Today a size catalogue shared across providers; now implementation + size + backend | Recreate classes per size and backend; map existing Machines by `spec.machineClass` + `spec.provider` + `vitistack.io/kubevirt-config` |
| 4 | **Machine names / NodePool adoption** | Names are node names and hostnames (`<clusterId>-wrk<i>`) | NodePool adopts by label; keep index-based names for new Machines |
| 5 | **`clusterId` as identity** | Secrets, Machines, VIP objects and labels are keyed by `spec.data.clusterId` | Keep `clusterId` as identity |
| 6 | **Event-driven kubevirt** | VMs live on remote clusters (KubevirtConfig) without watches | Per-cluster cache with watches |
| 7 | **Upgrades in progress** | State spread over annotations and a JSON blob in the Secret | Block migration during upgrades, or import explicitly |

## 3. Gaps in the proposal

| # | Gap | Suggestion |
|---|---|---|
| 1 | ~~Size vs implementation~~ | Decided: `MachineClass` as proposed, linked to `MachineProvider` (main proposal §3.1, §5.1) |
| 2 | Business metadata (`datacenter`, `region`, `zone`, `project`, `workspace`, `workorder`, `environment`) | Keep as labels or a `spec.metadata` block |
| 3 | Pool storage, architecture, labels/annotations | Add to `MachineTemplateSpec` |
| 4 | ~~`KubernetesProvider` / `MachineProvider` registrations~~ | Decided: kept as metadata registrations sent upstream; referenced by classes (main proposal §3.1) |
| 5 | `spec.provider` pattern on NetworkConfiguration / ControlPlaneVirtualSharedIP (kea, static-ip, nms operators) | Accept mixed patterns, or plan classes there too |
| 6 | VIP objects have no owner reference; deleted manually in `performCleanup` | Include them in the ownership model |
