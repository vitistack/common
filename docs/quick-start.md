# Vitistack Quick Start Guide

This guide will help you get started with Vitistack CRDs for infrastructure management.

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3.x (for installation)

## Installation

### Install CRDs

```bash
helm install vitistack-crds oci://ghcr.io/vitistack/vitistack-crds --version 0.1.0
```

Verify installation:

```bash
kubectl get crds | grep vitistack.io
```

You should see:

```
controlplanevirtualsharedips.vitistack.io
kubernetesclusters.vitistack.io
kubernetesproviders.vitistack.io
kubevirtconfigs.vitistack.io
machineproviders.vitistack.io
machines.vitistack.io
networkconfigurations.vitistack.io
networknamespaces.vitistack.io
proxmoxconfigs.vitistack.io
vitistacks.vitistack.io
```

## First Steps

### 1. Create a MachineProvider

Define where your machines will be provisioned:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: vitistack.io/v1alpha1
kind: MachineProvider
metadata:
  name: my-proxmox
  namespace: default
spec:
  type: proxmox
  region: datacenter-1
  endpoint:
    url: https://proxmox.example.com:8006
  authentication:
    type: token
    credentialsRef:
      name: proxmox-token
      namespace: default
  storage:
    defaultStorageClass: "local-lvm"
EOF
```

### 2. Create Your First Machine

```bash
cat <<EOF | kubectl apply -f -
apiVersion: vitistack.io/v1alpha1
kind: Machine
metadata:
  name: test-vm
  namespace: default
spec:
  providerRef:
    name: my-proxmox
  instanceType: medium
  cpu:
    cores: 2
    sockets: 1
  memory: 4096
  disks:
    - name: root
      size: 20
      type: ssd
      boot: true
  network:
    interfaces:
      - name: eth0
        ipv4: "10.0.1.100/24"
        gateway: "10.0.1.1"
  os:
    type: linux
    distribution: ubuntu
    version: "22.04"
  sshKeys:
    - "ssh-rsa AAAA... your-key-here"
EOF
```

### 3. Check Machine Status

```bash
kubectl get machines
kubectl describe machine test-vm
```

### 4. Create a Network Configuration

```bash
cat <<EOF | kubectl apply -f -
apiVersion: vitistack.io/v1alpha1
kind: NetworkConfiguration
metadata:
  name: my-network
  namespace: default
spec:
  cidr: "10.0.0.0/16"
  gateway: "10.0.0.1"
  vlan:
    id: 100
  subnets:
    - name: web
      cidr: "10.0.1.0/24"
    - name: app
      cidr: "10.0.2.0/24"
  dns:
    servers:
      - 8.8.8.8
      - 8.8.4.4
EOF
```

## Creating a Kubernetes Cluster

### 1. Create a KubernetesProvider

```bash
cat <<EOF | kubectl apply -f -
apiVersion: vitistack.io/v1alpha1
kind: KubernetesProvider
metadata:
  name: my-k8s-provider
  namespace: default
spec:
  type: kubeadm
  version: "1.28.0"
  machineProviderRef:
    name: my-proxmox
  cluster:
    version: "1.28.0"
    dns:
      type: CoreDNS
  network:
    cni: calico
    podCIDR: "192.168.0.0/16"
    serviceCIDR: "10.96.0.0/12"
  addons:
    ingressController:
      enabled: true
      type: nginx
EOF
```

### 2. Create a KubernetesCluster

```bash
cat <<EOF | kubectl apply -f -
apiVersion: vitistack.io/v1alpha1
kind: KubernetesCluster
metadata:
  name: my-cluster
  namespace: default
spec:
  data:
    provider: vitistack
    environment: development
  topology:
    controlPlane:
      replicas: 3
      machineClass: medium
    workers:
      nodePools:
        - name: general
          replicas: 3
          machineClass: large
          autoscaling:
            enabled: true
            minReplicas: 2
            maxReplicas: 10
EOF
```

### 3. Monitor Cluster Creation

```bash
kubectl get kubernetesclusters
kubectl describe kubernetescluster my-cluster
```

## Common Commands

### List All Resources

```bash
# Machines
kubectl get machines -A

# Machine Providers
kubectl get machineproviders -A

# Networks
kubectl get networkconfigurations -A

# Kubernetes Clusters
kubectl get kubernetesclusters -A
```

### Using Short Names

```bash
# Machines
kubectl get m

# Machine Providers
kubectl get mp

# Kubernetes Clusters
kubectl get kc

# Kubernetes Providers
kubectl get kp
```

### Watch Resources

```bash
kubectl get machines -w
kubectl get kc -w
```

### Delete Resources

```bash
kubectl delete machine test-vm
kubectl delete kubernetescluster my-cluster
```

## Example Workflows

### Provision a Web Server

```yaml
apiVersion: vitistack.io/v1alpha1
kind: Machine
metadata:
  name: web-01
spec:
  providerRef:
    name: my-proxmox
  instanceType: medium
  cpu:
    cores: 4
  memory: 8192
  disks:
    - name: root
      size: 50
      boot: true
  network:
    interfaces:
      - name: eth0
        ipv4: "10.0.1.10/24"
        gateway: "10.0.1.1"
  os:
    type: linux
    distribution: ubuntu
    version: "22.04"
  tags:
    role: webserver
    environment: production
  backup:
    enabled: true
    schedule: "0 2 * * *"
    retentionDays: 7
```

### Create a Development Cluster

```yaml
apiVersion: vitistack.io/v1alpha1
kind: KubernetesCluster
metadata:
  name: dev-cluster
spec:
  data:
    environment: development
  topology:
    controlPlane:
      replicas: 1
      machineClass: small
    workers:
      nodePools:
        - name: dev-pool
          replicas: 2
          machineClass: medium
```

## Troubleshooting

### Check CRD Status

```bash
kubectl get crds | grep vitistack
kubectl describe crd machines.vitistack.io
```

### View Resource Events

```bash
kubectl get events --field-selector involvedObject.name=test-vm
```

### Check Controller Logs

```bash
kubectl logs -n vitistack-system -l app=vitistack-controller --tail=100 -f
```

### Validate Resource Before Creating

```bash
kubectl apply --dry-run=client -f machine.yaml
kubectl apply --dry-run=server -f machine.yaml
```

## Next Steps

- Read the [full CRD documentation](crds.md)
- Check out [examples directory](../examples/) for more samples
- Learn about [networking concepts](networking.md)
- Understand [provider configuration](providers.md)

## Getting Help

- GitHub Issues: https://github.com/vitistack/common/issues
- Documentation: https://github.com/vitistack/common/tree/main/docs
- Examples: https://github.com/vitistack/common/tree/main/examples
