# Vitistack CRDs Helm Chart

This Helm chart installs the Vitistack Custom Resource Definitions (CRDs).

## Installation

### Install from OCI Registry (Recommended)

```bash
# Install latest version
helm install vitistack-crds oci://ghcr.io/vitistack/vitistack-crds

# Install specific version
helm install vitistack-crds oci://ghcr.io/vitistack/vitistack-crds --version 0.1.0
```

### Install from source

```bash
helm install vitistack-crds ./charts/vitistack-crds
```

### Install with custom namespace

```bash
# From OCI registry
helm install vitistack-crds oci://ghcr.io/vitistack/vitistack-crds \
  --version 0.1.0 \
  --create-namespace \
  --namespace vitistack-system

# From source
helm install vitistack-crds ./charts/vitistack-crds \
  --create-namespace \
  --namespace vitistack-system
```

### Pull and inspect before installing

```bash
# Pull the chart from OCI registry
helm pull oci://ghcr.io/vitistack/vitistack-crds --version 0.1.0

# Install from downloaded package
helm install vitistack-crds vitistack-crds-0.1.0.tgz
```

## Upgrading

```bash
# From OCI registry
helm upgrade vitistack-crds oci://ghcr.io/vitistack/vitistack-crds --version 0.1.0

# From source
helm upgrade vitistack-crds ./charts/vitistack-crds
```

## Uninstallation

```bash
helm uninstall vitistack-crds
```

**Note:** By default, CRDs are kept when uninstalling to prevent data loss. To remove them manually:

```bash
kubectl delete crd vitistacks.vitistack.io
kubectl delete crd kubernetesclusters.vitistack.io
kubectl delete crd kubernetesproviders.vitistack.io
kubectl delete crd kubevirtconfigs.vitistack.io
kubectl delete crd loadbalancers.vitistack.io
kubectl delete crd machineproviders.vitistack.io
kubectl delete crd machines.vitistack.io
kubectl delete crd networkconfigurations.vitistack.io
kubectl delete crd networknamespaces.vitistack.io
kubectl delete crd proxmoxconfigs.vitistack.io
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter     | Description                             | Default |
| ------------- | --------------------------------------- | ------- |
| `crds.keep`   | Keep CRDs on helm uninstall             | `true`  |
| `annotations` | Annotations to add to all CRD resources | `{}`    |
| `labels`      | Labels to add to all CRD resources      | `{}`    |

## CRDs Included

This chart installs the following CRDs:

- `vitistacks.vitistack.io`
- `kubernetesclusters.vitistack.io`
- `kubernetesproviders.vitistack.io`
- `kubevirtconfigs.vitistack.io`
- `loadbalancers.vitistack.io`
- `machineproviders.vitistack.io`
- `machines.vitistack.io`
- `networkconfigurations.vitistack.io`
- `networknamespaces.vitistack.io`
- `proxmoxconfigs.vitistack.io`

## Notes

- CRDs are cluster-scoped resources and will be available across all namespaces
- The chart uses the `"helm.sh/resource-policy": keep` annotation by default to preserve CRDs and their custom resources on chart deletion
