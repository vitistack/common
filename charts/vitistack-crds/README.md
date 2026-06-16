# Vitistack CRDs Helm Chart

This Helm chart installs the Vitistack Custom Resource Definitions (CRDs) and a
CRD **conversion webhook** that serves `v1alpha1 <-> v1alpha2` conversion for
`NetworkNamespace` resources. The webhook is deployed alongside the CRDs so any
combination of operators can rely on conversion without depending on a specific
operator.

## Prerequisites

- A Kubernetes cluster.
- The conversion webhook requires [cert-manager](https://cert-manager.io/)
  (including its `cainjector`) installed in the cluster **before** this chart.
  cert-manager provisions the webhook's TLS certificate and injects the CA bundle
  into the CRD's `spec.conversion`. If you do not run cert-manager, install CRDs
  only by disabling the webhook: `--set conversionWebhook.enabled=false`.

## Installing cert-manager (required dependency)

This chart does **not** bundle cert-manager — it is cluster-wide infrastructure
and should be managed as its own release so it can be shared by other workloads.
Install it once per cluster before installing (or before the first sync of) this
chart. Two reasons it is a hard dependency when the webhook is enabled:

1. It issues the webhook's serving certificate into the `…-conversion-webhook-tls`
   Secret. Without it the webhook pod cannot load its TLS cert and stays in
   `CrashLoopBackOff`.
2. Its `cainjector` populates `spec.conversion.webhook.clientConfig.caBundle` on
   the `NetworkNamespace` CRD (via the `cert-manager.io/inject-ca-from`
   annotation). Without it the API server will not trust the webhook and every
   `NetworkNamespace` operation fails.

### Manual install (Helm)

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true        # installs the cert-manager CRDs

# Wait until all three components (controller, webhook, cainjector) are Ready
kubectl -n cert-manager rollout status deploy/cert-manager
kubectl -n cert-manager rollout status deploy/cert-manager-webhook
kubectl -n cert-manager rollout status deploy/cert-manager-cainjector
```

Then install this chart as shown under [Installation](#installation).

### ArgoCD

Run cert-manager as a **separate** ArgoCD `Application` that syncs **before**
this chart. Use a lower `argocd.argoproj.io/sync-wave` on the cert-manager
Application (e.g. `"-1"`) than on the CRDs Application (e.g. `"0"`) so the
issuer/cainjector exist before the CRD references them.

```yaml
# cert-manager Application (sync-wave -1)
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cert-manager
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
spec:
  project: default
  source:
    repoURL: https://charts.jetstack.io
    chart: cert-manager
    targetRevision: v1.x.x
    helm:
      values: |
        crds:
          enabled: true
  destination:
    server: https://kubernetes.default.svc
    namespace: cert-manager
  syncPolicy:
    automated: { prune: true, selfHeal: true }
    syncOptions:
      - CreateNamespace=true
```

**Gotcha — perpetual `OutOfSync`:** cainjector mutates the CRD's `caBundle`
after ArgoCD syncs it, so ArgoCD reports the CRDs Application as drifted forever.
Tell ArgoCD to ignore that field on the CRDs Application:

```yaml
# CRDs Application (sync-wave 0)
spec:
  ignoreDifferences:
    - group: apiextensions.k8s.io
      kind: CustomResourceDefinition
      name: networknamespaces.vitistack.io
      jqPathExpressions:
        - .spec.conversion.webhook.clientConfig.caBundle
```

## Installation

### Install from OCI Registry (Recommended)

```bash
# Install latest version
helm install vitistack-crds oci://ghcr.io/vitistack/helm/crds

# Install specific version
helm install vitistack-crds oci://ghcr.io/vitistack/helm/crds --version 1.0.0
```

### Install from source

```bash
helm install vitistack-crds ./charts/vitistack-crds
```

### Install with custom namespace

```bash
# From OCI registry
helm install vitistack-crds oci://ghcr.io/vitistack/helm/crds \
  --version 1.0.0 \
  --create-namespace \
  --namespace vitistack-system

# From source
helm install vitistack-crds ./charts/vitistack-crds \
  --create-namespace \
  --namespace vitistack-system
```

### Install CRDs only (no conversion webhook)

```bash
helm install vitistack-crds oci://ghcr.io/vitistack/helm/crds \
  --set conversionWebhook.enabled=false
```

### Pull and inspect before installing

```bash
# Pull the chart from OCI registry
helm pull oci://ghcr.io/vitistack/helm/crds --version 1.0.0

# Install from downloaded package
helm install vitistack-crds crds-1.0.0.tgz
```

## Upgrading

```bash
# From OCI registry
helm upgrade vitistack-crds oci://ghcr.io/vitistack/helm/crds --version 1.0.0

# From source
helm upgrade vitistack-crds ./charts/vitistack-crds
```

## Uninstallation

```bash
helm uninstall vitistack-crds
```

**Note:** By default, CRDs are kept when uninstalling to prevent data loss. To remove them manually:

```bash
kubectl delete crd clusterstorageclasses.vitistack.io
kubectl delete crd clusterstorages.vitistack.io
kubectl delete crd controlplanevirtualsharedips.vitistack.io
kubectl delete crd etcdbackups.vitistack.io
kubectl delete crd ipallocations.vitistack.io
kubectl delete crd kubernetesclusters.vitistack.io
kubectl delete crd kubernetesproviders.vitistack.io
kubectl delete crd kubevirtconfigs.vitistack.io
kubectl delete crd machineclasses.vitistack.io
kubectl delete crd machineproviders.vitistack.io
kubectl delete crd machines.vitistack.io
kubectl delete crd networkconfigurations.vitistack.io
kubectl delete crd networknamespaces.vitistack.io
kubectl delete crd proxmoxconfigs.vitistack.io
kubectl delete crd vitistacks.vitistack.io
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter                                | Description                                                                          | Default                                          |
| ---------------------------------------- | ------------------------------------------------------------------------------------ | ------------------------------------------------ |
| `crds.keep`                              | Keep CRDs on helm uninstall                                                          | `true`                                           |
| `annotations`                            | Annotations to add to all CRD resources                                             | `{}`                                             |
| `labels`                                 | Labels to add to all CRD resources                                                  | `{}`                                             |
| `conversionWebhook.enabled`              | Deploy the `v1alpha1 <-> v1alpha2` conversion webhook (and wire the CRD to it)       | `true`                                           |
| `conversionWebhook.image.repository`     | Webhook image repository                                                            | `ghcr.io/vitistack/viti-crd-conversion-webhook`  |
| `conversionWebhook.image.tag`            | Webhook image tag (empty defaults to the chart `appVersion`)                         | `""`                                             |
| `conversionWebhook.image.pullPolicy`     | Webhook image pull policy                                                            | `IfNotPresent`                                   |
| `conversionWebhook.replicas`             | Number of webhook replicas                                                          | `2`                                              |
| `conversionWebhook.port`                 | Webhook server port (HTTPS)                                                          | `9443`                                           |
| `conversionWebhook.resources`            | Webhook container resource requests/limits                                          | see `values.yaml`                                |
| `conversionWebhook.certManager.enabled`  | Provision the webhook TLS certificate via cert-manager                              | `true`                                           |
| `conversionWebhook.certManager.issuerRef`| Use an existing `Issuer`/`ClusterIssuer`; if empty, a self-signed `Issuer` is created | `{}`                                             |

## Conversion Webhook

`NetworkNamespace` is served at both `v1alpha1` and `v1alpha2` (`v1alpha2` is the
storage version). When `conversionWebhook.enabled` is `true` (the default) the
chart deploys an HTTPS conversion webhook and wires the CRD's
`spec.conversion.strategy: Webhook` to it.

Resources created when enabled:

- **Deployment** (`<release>-conversion-webhook`) — runs the standalone webhook
  server, `2` replicas by default.
- **Service** (`<release>-conversion-webhook`) — `443 -> 9443`.
- **Certificate** + **Issuer** — a cert-manager `Certificate`; when
  `certManager.issuerRef` is empty a self-signed `Issuer` is created and used.
- **ServiceAccount** — pod identity (the server makes no API calls, so no RBAC).

The CRD carries a `cert-manager.io/inject-ca-from` annotation so cert-manager's
cainjector populates `spec.conversion.webhook.clientConfig.caBundle`.

**Ordering note:** until the webhook pods are `Ready` and cert-manager has
injected the CA bundle, `NetworkNamespace` operations will fail — this is normal
for conversion webhooks and resolves once the pods come up.

To use an existing issuer instead of the self-signed fallback:

```bash
helm install vitistack-crds oci://ghcr.io/vitistack/helm/crds \
  --set conversionWebhook.certManager.issuerRef.name=my-ca \
  --set conversionWebhook.certManager.issuerRef.kind=ClusterIssuer
```

## CRDs Included

This chart installs the following CRDs:

- `clusterstorageclasses.vitistack.io`
- `clusterstorages.vitistack.io`
- `controlplanevirtualsharedips.vitistack.io`
- `etcdbackups.vitistack.io`
- `ipallocations.vitistack.io`
- `kubernetesclusters.vitistack.io`
- `kubernetesproviders.vitistack.io`
- `kubevirtconfigs.vitistack.io`
- `machineclasses.vitistack.io`
- `machineproviders.vitistack.io`
- `machines.vitistack.io`
- `networkconfigurations.vitistack.io`
- `networknamespaces.vitistack.io`
- `proxmoxconfigs.vitistack.io`
- `vitistacks.vitistack.io`

## Notes

- CRDs are cluster-scoped resources and will be available across all namespaces.
- The chart uses the `"helm.sh/resource-policy": keep` annotation by default to preserve CRDs and their custom resources on chart deletion.
