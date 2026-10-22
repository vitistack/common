# Vitistack Custom Resource Definitions (CRDs)

This document describes the Custom Resource Definitions (CRDs) provided by Vitistack for managing infrastructure and Kubernetes clusters.

## Overview

Vitistack provides a comprehensive set of CRDs for declarative infrastructure management:

- **Machine**: Virtual machine provisioning and management
- **MachineProvider**: Configuration for machine provisioning backends (Proxmox, KubeVirt, etc.)
- **MachineClass**: Machine size and resource presets
- **NetworkConfiguration**: Network topology and configuration
- **NetworkNamespace**: Network isolation and segmentation
- **KubernetesCluster**: Kubernetes cluster lifecycle management
- **KubernetesProvider**: Kubernetes cluster provider configuration
- **ControlPlaneVirtualSharedIP**: Shared IP management for control planes
- **EtcdBackup**: Etcd backup configuration and scheduling
- **VitiStack**: Complete infrastructure stack definition
- **ProxmoxConfig**: Proxmox-specific configuration
- **KubevirtConfig**: KubeVirt-specific configuration

## Installation

### Using Helm (Recommended)

```bash
helm install vitistack-crds oci://ghcr.io/vitistack/vitistack-crds --version 0.1.0
```

### Using kubectl

```bash
kubectl apply -f https://github.com/vitistack/common/releases/download/v0.1.0/crds.yaml
```

### From Source

```bash
git clone https://github.com/vitistack/common
cd common
make install-crds
```

## CRD Reference

### Machine

`machines.vitistack.io/v1alpha1`

Represents a virtual machine instance with full lifecycle management.

**Short Name**: `m`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: Machine
metadata:
  name: web-server-01
  namespace: default
spec:
  providerRef:
    name: proxmox-provider
    namespace: default
  instanceType: "medium" # or specific specs
  cpu:
    cores: 4
    sockets: 1
  memory: 8192 # MB
  disks:
    - name: root
      size: 50 # GB
      type: ssd
      boot: true
  network:
    interfaces:
      - name: eth0
        bridge: vmbr0
        ipv4: "10.0.1.100/24"
        gateway: "10.0.1.1"
  os:
    type: linux
    distribution: ubuntu
    version: "22.04"
  sshKeys:
    - ssh-rsa AAAA...
  tags:
    environment: production
    role: webserver
```

**Key Fields**:

- `spec.providerRef`: Reference to MachineProvider
- `spec.instanceType`: Machine size preset
- `spec.cpu`: CPU configuration (cores, sockets, threads)
- `spec.memory`: Memory in MB
- `spec.disks`: Storage configuration
- `spec.network`: Network interfaces and configuration
- `spec.os`: Operating system details
- `spec.sshKeys`: SSH public keys for access
- `status.phase`: Current lifecycle phase
- `status.state`: Running state (running, stopped, etc.)

### MachineProvider

`machineproviders.vitistack.io/v1alpha1`

Configures the backend provider for machine provisioning (Proxmox, KubeVirt, etc.).

**Short Name**: `mp`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: MachineProvider
metadata:
  name: proxmox-datacenter-1
  namespace: vitistack-system
spec:
  type: proxmox
  region: us-east-1
  zones:
    - zone-a
    - zone-b
  endpoint:
    url: https://proxmox.example.com:8006
    insecureSkipVerify: false
  authentication:
    type: token
    credentialsRef:
      name: proxmox-credentials
      namespace: vitistack-system
  capabilities:
    maxMachines: 100
    supportedInstanceTypes:
      - small
      - medium
      - large
    features:
      - snapshot
      - backup
      - migration
  storage:
    defaultStorageClass: "local-lvm"
    availableStorage:
      - name: local-lvm
        type: lvm
      - name: nfs-storage
        type: nfs
```

**Key Fields**:

- `spec.type`: Provider type (proxmox, kubevirt, etc.)
- `spec.endpoint`: Provider API endpoint
- `spec.authentication`: Authentication configuration
- `spec.capabilities`: Provider capabilities and limits
- `status.health`: Provider health status

### NetworkConfiguration

`networkconfigurations.vitistack.io/v1alpha1`

Describes the network interfaces a machine should be attached to and reports the
IP addresses that were allocated for those interfaces. A NetworkConfiguration
references a NetworkNamespace (by name) and is processed by the IP allocation
operator identified in `spec.provider` (e.g. `kea`, `static-ip-operator`).

**Short Name**: `nc`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: NetworkConfiguration
metadata:
  name: web-server-01-net
  namespace: default
spec:
  name: test-network
  description: Primary network configuration for web-server-01
  networkNamespaceName: tenant-acme
  provider: static-ip-operator # "kea" for DHCP, "static-ip-operator" for static
  datacenterIdentifier: no-west-az1
  supervisorIdentifier: p-west-mgmt
  clusterIdentifier: production
  networkInterfaces:
    - name: eth0
      macAddress: "52:54:00:12:34:56"
```

**Key Spec Fields**:

- `spec.name`: Unique name for the NetworkConfiguration (required, 2–32 chars, `[A-Za-z0-9_-]`)
- `spec.description`: Free-text description (optional, max 256 chars)
- `spec.networkNamespaceName`: Name of the NetworkNamespace to allocate from. If empty, the operator falls back to a list-based lookup (deprecated — emits a warning)
- `spec.provider`: Identifies which IP allocation operator handles this resource. Recommended when multiple allocation operators run in the same cluster. When empty, the operator inherits from the referenced NetworkNamespace's `ipAllocation.provider` (deprecated — emits a warning)
- `spec.datacenterIdentifier` / `supervisorIdentifier` / `clusterIdentifier`: Scoping identifiers (optional, same pattern rules as above)
- `spec.networkInterfaces[]`: Interfaces to attach. Each entry accepts `name`, `macAddress`, `vlan`, `ipv4Subnet`/`ipv6Subnet`, `ipv4Gateway`/`ipv6Gateway`, `dns[]`, and pre-seeded addresses

**Key Status Fields**:

- `status.phase` / `status.status` / `status.message`: High-level reconciliation state
- `status.networkInterfaces[]`: Per-interface allocation result, including:
  - `ipv4Addresses[]` / `ipv6Addresses[]`: Allocated addresses
  - `dhcpReserved`: Whether a reservation has been recorded on the DHCP server
  - `ipAllocated`: `true` once an IP has been successfully allocated, regardless of method
  - `allocationMethod`: `dhcp` or `static` — how this interface's address was assigned
  - `allocationExpiry`: Expiry timestamp for static allocations that carry a TTL (see `NetworkNamespace.spec.ipAllocation.static.ttlSeconds`)

### NetworkNamespace

`networknamespaces.vitistack.io/v1alpha1`

Represents a logical network segment within a datacenter/supervisor. A
NetworkNamespace declares how IP addresses are allocated (DHCP via an external
server such as Kea, or static allocation from a defined pool) and is referenced
by NetworkConfigurations that need addresses from it.

**Short Name**: `nn`

**Example — DHCP (default behavior)**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: NetworkNamespace
metadata:
  name: tenant-acme
  namespace: default
spec:
  datacenterIdentifier: no-west-az1
  supervisorIdentifier: tenant-acme
  # ipAllocation omitted → DHCP-based allocation (nms-operator + kea-operator)
```

**Example — static IP allocation**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: NetworkNamespace
metadata:
  name: test-network-static
  namespace: default
spec:
  datacenterIdentifier: no-west-az1
  supervisorIdentifier: p-west-mgmt
  ipAllocation:
    type: static
    provider: static-ip-operator
    static:
      ipv4CIDR: "10.100.1.0/24"
      ipv4Gateway: "10.100.1.1"
      ipv4RangeStart: "10.100.1.10" # optional, defaults to x.x.x.2
      ipv4RangeEnd: "10.100.1.200" # optional, defaults to last usable address
      vlanId: 100 # optional (0–4094); kubevirt-operator uses this to tag the NetworkAttachmentDefinition
      ttlSeconds: 3600 # optional, min 60, default 3600
      dns:
        - 8.8.8.8
        - 8.8.4.4
```

**Example — DHCP with Kea client class overrides**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: NetworkNamespace
metadata:
  name: tenant-acme-restricted
  namespace: default
spec:
  datacenterIdentifier: no-west-az1
  supervisorIdentifier: p-west-mgmt
  ipAllocation:
    type: dhcp
    provider: kea
    dhcp:
      requireClientClasses:
        - trusted-vms
```

**Key Spec Fields**:

- `spec.datacenterIdentifier`: Datacenter identifier (required, 2–32 chars, `[A-Za-z0-9_-]`), e.g. `no-west-az1`
- `spec.supervisorIdentifier`: Unique name per datacenter (required, same rules as above)
- `spec.ipAllocation`: IP allocation configuration (optional — omitting it preserves the legacy DHCP-based behavior)
  - `spec.ipAllocation.type`: `dhcp` or `static` (required when `ipAllocation` is set)
  - `spec.ipAllocation.provider`: Operator handling the allocation. Known values: `kea`, `static-ip-operator`
  - `spec.ipAllocation.static`: Required when `type: static`. Fields: `ipv4CIDR`, `ipv4Gateway` (required), `ipv4RangeStart`, `ipv4RangeEnd`, `vlanId`, `dns[]`, `ttlSeconds`
  - `spec.ipAllocation.dhcp.requireClientClasses[]`: Kea DHCP client classes that must match for lease allocation

**Key Status Fields**:

- `status.phase` / `status.status` / `status.message`: High-level reconciliation state
- `status.namespaceId`, `status.ipv4Prefix`, `status.ipv6Prefix`, `status.ipv4EgressIp`, `status.ipv6EgressIp`, `status.vlanId`: Observed network attributes
- `status.associatedKubernetesClusterIds[]`: Clusters that consume this namespace
- `status.ipAllocationStatus`: Current pool utilization (for both DHCP and static), including:
  - `type`, `provider`: Active allocation method and operator
  - `allocatedCount`, `availableCount`, `totalCount`: Pool utilization counters
  - `allocatedIPs[]`: Each entry reports `{ ip, networkConfiguration }` — the IP and the NetworkConfiguration that owns it

### KubernetesCluster

`kubernetesclusters.vitistack.io/v1alpha1`

Manages the full lifecycle of a Kubernetes cluster.

**Short Name**: `kc`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: KubernetesCluster
metadata:
  name: production-cluster
  namespace: default
spec:
  data:
    provider: vitistack
    region: us-east-1
    zone: zone-a
    environment: production
  topology:
    controlPlane:
      replicas: 3
      machineClass: medium
      metadata:
        labels:
          node-role: control-plane
    workers:
      nodePools:
        - name: general
          replicas: 5
          machineClass: large
          autoscaling:
            enabled: true
            minReplicas: 3
            maxReplicas: 10
          metadata:
            labels:
              workload: general
        - name: gpu
          replicas: 2
          machineClass: gpu-large
          metadata:
            labels:
              workload: gpu
```

**Key Fields**:

- `spec.topology.controlPlane`: Control plane configuration
- `spec.topology.workers.nodePools`: Worker node pools
- `status.state.cluster.resources`: Resource utilization
- `status.state.endpoints`: Cluster endpoints

### KubernetesProvider

`kubernetesproviders.vitistack.io/v1alpha1`

Configures how Kubernetes clusters are provisioned and managed.

**Short Name**: `kp`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: KubernetesProvider
metadata:
  name: vitistack-k8s
  namespace: vitistack-system
spec:
  type: kubeadm
  version: "1.28.0"
  region: us-east-1
  machineProviderRef:
    name: proxmox-provider
    namespace: vitistack-system
  cluster:
    version: "1.28.0"
    apiServer:
      admissionPlugins:
        - NodeRestriction
        - PodSecurityPolicy
    dns:
      type: CoreDNS
      upstreamServers:
        - 8.8.8.8
  network:
    cni: calico
    podCIDR: "192.168.0.0/16"
    serviceCIDR: "10.96.0.0/12"
  addons:
    ingressController:
      enabled: true
      type: nginx
    storage:
      enabled: true
      defaultClass: standard
```

### ControlPlaneVirtualSharedIP

`controlplanevirtualsharedips.vitistack.io/v1alpha1`

Manages shared virtual IPs for highly available control planes.

**Short Name**: `cpvip`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: ControlPlaneVirtualSharedIP
metadata:
  name: k8s-api-vip
  namespace: default
spec:
  ipAddress: "10.0.1.10"
  port: 6443
  poolMembers:
    - 10.0.1.11
    - 10.0.1.12
    - 10.0.1.13
```

### EtcdBackup

`etcdbackups.vitistack.io/v1alpha1`

Manages etcd backup configuration and scheduling for Kubernetes clusters.

**Short Name**: `eb`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: EtcdBackup
metadata:
  name: prod-cluster-backup
  namespace: default
spec:
  clusterName: production-cluster
  schedule: "0 */6 * * *" # Every 6 hours
  retention: 7
  storageLocation:
    type: s3
    bucket: my-etcd-backups
    path: prod-cluster/
    secretRef: backup-credentials
```

**Key Fields**:

- `spec.clusterName`: Name of the Kubernetes cluster to backup (required)
- `spec.schedule`: Cron schedule for automated backups (optional)
- `spec.retention`: Number of backups to retain (default: 7)
- `spec.storageLocation`: Storage destination configuration
  - `type`: Storage type (`s3`, `gcs`, `azure`, `local`)
  - `bucket`: Bucket name for cloud storage
  - `path`: Path/prefix within the storage location
  - `secretRef`: Reference to secret containing storage credentials
- `status.phase`: Current phase (`Pending`, `Running`, `Completed`, `Failed`)
- `status.lastBackupTime`: Timestamp of last successful backup
- `status.nextBackupTime`: Scheduled time for next backup
- `status.backupSize`: Size of the last backup
- `status.backupCount`: Current number of stored backups
- `status.conditions`: Standard Kubernetes conditions

**Storage Types**:

| Type    | Description                     | Required Fields       |
| ------- | ------------------------------- | --------------------- |
| `s3`    | AWS S3 or S3-compatible storage | `bucket`, `secretRef` |
| `gcs`   | Google Cloud Storage            | `bucket`, `secretRef` |
| `azure` | Azure Blob Storage              | `bucket`, `secretRef` |
| `local` | Local filesystem storage        | `path`                |

### VitiStack

`vitistacks.vitistack.io/v1alpha1`

Complete infrastructure stack definition combining all components. This is a cluster-scoped resource that represents a datacenter or availability zone.

**Short Name**: `vs`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: Vitistack
metadata:
  name: production-stack
spec:
  displayName: Production Stack
  region: south
  zone: az1
  infrastructure: prod
  description: Production infrastructure stack
  location:
    country: "no"
    city: Oslo
  machineProviders:
    - name: proxmox-provider
      namespace: vitistack-system
      priority: 1
      enabled: true
  kubernetesProviders:
    - name: k8s-provider
      namespace: vitistack-system
  networking:
    vpcs:
      - name: production-vpc
        cidr: "10.0.0.0/16"
        default: true
        subnets:
          - name: web-tier
            cidr: "10.0.1.0/24"
            zone: az1
    dns:
      domain: prod.example.com
      servers:
        - 8.8.8.8
    ipAllocation:
      - zone: "" # empty = default for all zones without their own entry
        providers:
          - type: dhcp
            enabled: true
            default: true
          - type: static
            enabled: true
      - zone: az1
        providers:
          - type: static
            enabled: true
            default: true
            configuration:
              operator: static-ip-operator
  backup:
    enabled: true
    schedule: "0 2 * * *"
    retentionPolicy:
      daily: 7
      weekly: 4
      monthly: 12
  resourceQuotas:
    maxMachines: 100
    maxClusters: 10
    maxCPUCores: 500
    maxMemoryGB: 2048
```

**Key Spec Fields**:

- `spec.displayName`: Human-readable name for the vitistack (required)
- `spec.region`: Geographical region (required)
- `spec.zone`: Availability zone within the region
- `spec.infrastructure`: Environment type (e.g., prod, test, mgmt)
- `spec.description`: Additional context about the vitistack
- `spec.location`: Detailed location information (country, city, coordinates)
- `spec.machineProviders`: List of machine provider references
- `spec.kubernetesProviders`: List of Kubernetes provider references
- `spec.networking`: Network configuration (VPCs, subnets, DNS, firewall, IP allocation)
- `spec.networking.ipAllocation[]`: Per-zone IP allocation providers available in this vitistack. Each entry has a `zone` (empty = default fallback) and a `providers[]` list. Each provider entry sets `type` (`dhcp`|`static`), `enabled` (default `true`), `default` (marks the provider a NetworkNamespace picks when it doesn't set its own), and optional `configuration` key/value settings
- `spec.security`: Security policies (encryption, access control, audit logging)
- `spec.monitoring`: Monitoring configuration
- `spec.backup`: Backup and disaster recovery policies
- `spec.resourceQuotas`: Resource limits for the vitistack

**Key Status Fields**:

- `status.phase`: Current phase (`Initializing`, `Provisioning`, `Ready`, `Degraded`, `Deleting`, `Failed`)
- `status.displayName`: Observed display name
- `status.region`: Observed region
- `status.zone`: Observed zone
- `status.infrastructure`: Observed infrastructure type
- `status.description`: Observed description
- `status.location`: Observed location information
- `status.machineProviders`: Discovered MachineProvider objects in the cluster
- `status.kubernetesProviders`: Discovered KubernetesProvider objects in the cluster
- `status.clusters`: Discovered KubernetesCluster objects in the vitistack
- `status.machineProviderCount`: Number of active machine providers
- `status.kubernetesProviderCount`: Number of active Kubernetes providers
- `status.activeMachines`: Number of active machines
- `status.activeClusters`: Number of active Kubernetes clusters
- `status.resourceUsage`: Current resource utilization (CPU, memory, storage)
- `status.conditions`: Standard Kubernetes conditions

**Discovered Provider Structure**:

```yaml
status:
  machineProviders:
    - name: proxmox-provider
      namespace: vitistack-system
      providerType: proxmox
      region: south
      zone: az1
      ready: true
      discoveredAt: "2025-12-04T12:00:00Z"
```

**Discovered Cluster Structure**:

```yaml
status:
  clusters:
    - name: prod-cluster
      namespace: default
      phase: Running
      version: "1.28.0"
      controlPlaneReplicas: 3
      workerReplicas: 5
      ready: true
      discoveredAt: "2025-12-04T12:00:00Z"
```

### ProxmoxConfig

`proxmoxconfigs.vitistack.io/v1alpha1`

Proxmox-specific configuration for virtual machine provisioning.

**Short Name**: `pxc`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: ProxmoxConfig
metadata:
  name: proxmox-defaults
  namespace: vitistack-system
spec:
  defaultNode: pve1
  defaultStorage: local-lvm
  defaultBridge: vmbr0
```

### KubevirtConfig

`kubevirtconfigs.vitistack.io/v1alpha1`

KubeVirt-specific configuration for VM provisioning on Kubernetes.

**Short Name**: `kvc`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: KubevirtConfig
metadata:
  name: kubevirt-defaults
  namespace: vitistack-system
spec:
  defaultNamespace: default
  defaultStorageClass: local-path
```

## Common Patterns

### Referencing Resources

Most resources support references to other resources using a standard format:

```yaml
someRef:
  name: resource-name
  namespace: resource-namespace # optional, defaults to same namespace
```

### Labels and Selectors

Use standard Kubernetes labels for organization:

```yaml
metadata:
  labels:
    environment: production
    tier: backend
    app: myapp
```

### Status Conditions

All resources follow Kubernetes conventions for status conditions:

```yaml
status:
  conditions:
    - type: Ready
      status: "True"
      reason: ResourceReady
      message: "Resource is ready"
      lastTransitionTime: "2025-11-12T10:00:00Z"
```

## Validation

All CRDs include OpenAPI v3 schema validation. Common validations include:

- **Required fields**: Enforced at API level
- **Format validation**: Email, CIDR, URL patterns
- **Range validation**: Min/max for numeric fields
- **Enum validation**: Restricted value sets

## RBAC

Typical RBAC setup for managing Vitistack resources:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: vitistack-admin
rules:
  - apiGroups: ["vitistack.io"]
    resources: ["*"]
    verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: vitistack-admin-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: vitistack-admin
subjects:
  - kind: User
    name: admin@example.com
```

## Troubleshooting

### Check CRD Installation

```bash
kubectl get crds | grep vitistack.io
```

### View Resource Status

```bash
kubectl get machines -A
kubectl describe machine <name> -n <namespace>
```

### Check Controller Logs

```bash
kubectl logs -n vitistack-system -l app=vitistack-controller
```

## Next Steps

- See [examples/](../examples/) for complete working examples
- Check the [API Reference](./api-reference.md) for detailed field documentation
- Review [Architecture](./architecture.md) for system design details

## Contributing

To add or modify CRDs:

1. Edit the Go types in `pkg/v1alpha1/`
2. Run `make generate` to regenerate CRDs and deepcopy code
3. Run `make verify-crds` to ensure format compliance
4. Update this documentation

For more details, see [CONTRIBUTING.md](../CONTRIBUTING.md).
