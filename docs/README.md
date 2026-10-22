# Vitistack Documentation

Welcome to the Vitistack documentation!

## Getting Started

- **[Quick Start Guide](quick-start.md)** - Get up and running in 5 minutes
- **[CRD Reference](crds.md)** - Complete reference for all Custom Resource Definitions

## Core Concepts

### Custom Resource Definitions (CRDs)

Vitistack provides the following CRDs for infrastructure management:

| CRD                         | Purpose                        | Short Name |
| --------------------------- | ------------------------------ | ---------- |
| Machine                     | Virtual machine provisioning   | `m`        |
| MachineProvider             | Machine provider configuration | `mp`       |
| NetworkConfiguration        | Network topology               | `netconf`  |
| NetworkNamespace            | Network isolation              | `netns`    |
| KubernetesCluster           | K8s cluster management         | `kc`       |
| KubernetesProvider          | K8s provider config            | `kp`       |
| ControlPlaneVirtualSharedIP | Control plane VIP              | `cpvip`    |
| VitiStack                   | Complete infrastructure stack  | `vs`       |
| ProxmoxConfig               | Proxmox configuration          | `pxc`      |
| KubevirtConfig              | KubeVirt configuration         | `kvc`      |

## Documentation Index

### User Guides

- [Quick Start Guide](quick-start.md) - First steps with Vitistack
- [CRD Reference](crds.md) - Detailed CRD documentation
- [Go Libraries](go-libraries.md) - Complete guide to all Go packages

### Examples

See the [examples directory](../examples/) for working examples.

### API Reference

- [Go Package Documentation](https://pkg.go.dev/github.com/vitistack/common) - Full API reference
- [CRD API Types](../pkg/v1alpha1/) - Go type definitions for CRDs

## Installation

### Install CRDs via Helm

```bash
helm install vitistack-crds oci://ghcr.io/vitistack/vitistack-crds --version 0.1.0
```

### Install CRDs via kubectl

```bash
kubectl apply -f https://github.com/vitistack/common/releases/latest/download/crds.yaml
```

### Use Go Library

```bash
go get github.com/vitistack/common@latest
```

## Common Use Cases

### 1. Provision Virtual Machines

Use the `Machine` CRD to declaratively manage VMs across multiple providers.

```bash
kubectl apply -f examples/machine-basic.yaml
```

### 2. Create Kubernetes Clusters

Use `KubernetesCluster` to provision and manage K8s clusters.

```bash
kubectl apply -f examples/kubernetes-cluster.yaml
```

### 3. Manage Networks

Use `NetworkConfiguration` to define network topology and VLANs.

```bash
kubectl apply -f examples/network-config.yaml
```

## Architecture

Vitistack follows a declarative, Kubernetes-native approach:

1. **CRDs** define the desired state of infrastructure
2. **Controllers** watch CRDs and reconcile actual state
3. **Providers** interact with underlying infrastructure (Proxmox, KubeVirt, etc.)

```
┌─────────────┐
│   kubectl   │
└──────┬──────┘
       │
       v
┌─────────────┐
│ Kubernetes  │
│   API       │
└──────┬──────┘
       │
       v
┌─────────────┐     ┌──────────────┐
│   CRDs      │────>│ Controllers  │
└─────────────┘     └──────┬───────┘
                           │
                           v
                    ┌──────────────┐
                    │  Providers   │
                    │ (Proxmox,    │
                    │  KubeVirt)   │
                    └──────────────┘
```

## Development

### Prerequisites

- Go 1.21+
- kubectl
- Access to a Kubernetes cluster

### Build from Source

```bash
git clone https://github.com/vitistack/common
cd common
make build
```

### Run Tests

```bash
make test
```

### Generate CRDs

```bash
make generate
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for details.

### Adding New CRDs

1. Define Go types in `pkg/v1alpha1/`
2. Add appropriate kubebuilder markers
3. Run `make generate`
4. Update documentation
5. Add examples

## Support

- **Issues**: https://github.com/vitistack/common/issues
- **Discussions**: https://github.com/vitistack/common/discussions
- **Documentation**: https://github.com/vitistack/common/tree/main/docs

## License

See [LICENSE](../LICENSE) for details.
