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

Defines network topology, VLANs, subnets, and routing.

**Short Name**: `netconf`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: NetworkConfiguration
metadata:
  name: production-network
  namespace: default
spec:
  cidr: "10.0.0.0/16"
  gateway: "10.0.0.1"
  vlan:
    id: 100
    name: production
  subnets:
    - name: web-tier
      cidr: "10.0.1.0/24"
      gateway: "10.0.1.1"
    - name: app-tier
      cidr: "10.0.2.0/24"
      gateway: "10.0.2.1"
  dns:
    servers:
      - 8.8.8.8
      - 8.8.4.4
    searchDomains:
      - example.com
```

### NetworkNamespace

`networknamespaces.vitistack.io/v1alpha1`

Provides network isolation and segmentation for multi-tenancy.

**Short Name**: `netns`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: NetworkNamespace
metadata:
  name: tenant-acme
  namespace: default
spec:
  cidr: "10.100.0.0/16"
  isolation: strict
  allowedNamespaces:
    - default
    - monitoring
```

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

Complete infrastructure stack definition combining all components.

**Short Name**: `vs`

**Example**:

```yaml
apiVersion: vitistack.io/v1alpha1
kind: VitiStack
metadata:
  name: production-stack
  namespace: default
spec:
  region: us-east-1
  environment: production
  components:
    - type: kubernetes-cluster
      name: prod-k8s
      config:
        version: "1.28.0"
        nodeCount: 5
    - type: network
      name: prod-network
      config:
        cidr: "10.0.0.0/16"
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
