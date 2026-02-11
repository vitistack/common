# Vitistack Common

[![Build and Test](https://github.com/vitistack/common/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/vitistack/common/actions/workflows/build-and-test.yml)
[![Security Scan](https://github.com/vitistack/common/actions/workflows/security-scan.yml/badge.svg)](https://github.com/vitistack/common/actions/workflows/security-scan.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitistack/common)](https://goreportcard.com/report/github.com/vitistack/common)
[![Go Reference](https://pkg.go.dev/badge/github.com/vitistack/common.svg)](https://pkg.go.dev/github.com/vitistack/common)

Shared infrastructure management components for Kubernetes-native platforms.

## What's Included

### 🎯 Custom Resource Definitions (CRDs)

Kubernetes-native APIs for declarative infrastructure management:

- **Machine** - Virtual machine provisioning and lifecycle
- **MachineProvider** - Multi-cloud provider configuration (Proxmox, KubeVirt)
- **MachineClass** - Machine size and resource presets
- **NetworkConfiguration** - Network topology and VLANs
- **NetworkNamespace** - Network isolation and segmentation
- **KubernetesCluster** - Kubernetes cluster management
- **KubernetesProvider** - Kubernetes cluster provider configuration
- **ControlPlaneVirtualSharedIP** - Shared IP management for control planes
- **EtcdBackup** - Etcd backup configuration and scheduling
- **VitiStack** - Infrastructure stack definition with auto-discovery of providers and clusters
- **ProxmoxConfig** - Proxmox-specific configuration
- **KubevirtConfig** - KubeVirt-specific configuration

📖 **[View full CRD documentation →](docs/crds.md)**

### 📦 Go Libraries

Small, focused libraries for cloud-native applications:

- **`vlog`** - Structured logging with Zap and logr adapter
- **`serialize`** - JSON helpers for quick serialization
- **`k8sclient`** - Kubernetes client initialization
- **`S3client`** - General s3 client
- **`crdcheck`** - CRD prerequisite validation
- **`dotenv`** - Smart environment configuration

📖 **[View Go library documentation →](docs/go-libraries.md)**

## Quick Start

### Install CRDs

```bash
# Using Helm (recommended)
# First, login to GitHub Container Registry
# Username: your GitHub username
# Password: a Personal Access Token (PAT) with `read:packages` scope
# Create a PAT at: https://github.com/settings/tokens/new?scopes=read:packages
helm registry login ghcr.io

helm install vitistack-crds oci://ghcr.io/vitistack/helm/crds

# Or using kubectl (no authentication required)
kubectl apply -f https://github.com/vitistack/common/releases/latest/download/crds.yaml
```

### Use Go Libraries


```bash
go get github.com/vitistack/common@latest
```

#### k8sclient

```go
import (
    "github.com/vitistack/common/pkg/loggers/vlog"
    "github.com/vitistack/common/pkg/clients/k8sclient"
)

func main() {
    vlog.Setup(vlog.Options{Level: "info", JSON: true})
    defer vlog.Sync()

    k8sclient.Init()
    vlog.Info("connected to kubernetes")
}
```

#### s3Client

```go
package main

import (
    "github.com/vitistack/common/pkg/loggers/vlog"
    "github.com/vitistack/common/pkg/clients/s3client/s3interface"
    "github.com/vitistack/common/pkg/clients/s3client/s3minioclient"
)

func main() {
  vlog.Setup(vlog.Options{Level: "info", JSON: true})
  defer vlog.Sync()

  s3, err := s3minioclient.NewS3Client(
	  s3interface.WithAccessKey("accesskey"),
	  s3interface.WithSecretKey("secretkey"),
	  s3interface.WithEndpoint("your-endpoint"),
	  s3interface.WithRegion("your region"),
	  s3interface.WithSecure(false),
  )
  vlog.Info("s3 client created")
}
```

## Documentation

| Resource              | Link                                                         |
| --------------------- | ------------------------------------------------------------ |
| 🚀 Quick Start Guide  | [docs/quick-start.md](docs/quick-start.md)                   |
| 📚 CRD Reference      | [docs/crds.md](docs/crds.md)                                 |
| 💻 Go Libraries       | [docs/go-libraries.md](docs/go-libraries.md)                 |
| 📖 Full Documentation | [docs/](docs/)                                               |
| 🔍 API Reference      | [pkg.go.dev](https://pkg.go.dev/github.com/vitistack/common) |
| 💡 Examples           | [examples/](examples/)                                       |

## Examples

### Provision a Virtual Machine

```yaml
apiVersion: vitistack.io/v1alpha1
kind: Machine
metadata:
  name: web-server
spec:
  providerRef:
    name: my-proxmox
  instanceType: medium
  cpu:
    cores: 4
  memory: 8192
  os:
    distribution: ubuntu
    version: "22.04"
```

### Create a Kubernetes Cluster

```yaml
apiVersion: vitistack.io/v1alpha1
kind: KubernetesCluster
metadata:
  name: production
spec:
  topology:
    controlPlane:
      replicas: 3
      machineClass: medium
    workers:
      nodePools:
        - name: general
          replicas: 5
          autoscaling:
            enabled: true
            minReplicas: 3
            maxReplicas: 10
```

### Use Go Libraries

```go
package main

import (
    "github.com/vitistack/common/pkg/loggers/vlog"
    "github.com/vitistack/common/pkg/settings/dotenv"
)

func main() {
    dotenv.LoadDotEnv()
    vlog.Setup(vlog.Options{Level: "info"})
    defer vlog.Sync()

    vlog.With("app", "myapp").Info("starting application")
}
```

## Development

```bash
# Clone repository
git clone https://github.com/vitistack/common
cd common

# Install dependencies
make deps

# Run tests
make test

# Generate CRDs
make generate

# Build
make build

# Lint code
make lint
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Adding New Features

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/vitistack/common/issues)
- 💬 [Discussions](https://github.com/vitistack/common/discussions)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ by the Vitistack Team**
