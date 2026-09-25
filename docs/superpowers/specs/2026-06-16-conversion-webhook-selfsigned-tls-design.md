# Conversion Webhook — Pluggable TLS Provider (cert-manager or self-signed)

**Date:** 2026-06-16
**Chart:** `common/charts/vitistack-crds`
**Status:** Approved design, pending implementation plan

## Problem

The `vitistack-crds` chart deploys a CRD conversion webhook for `NetworkNamespace`
(`v1alpha1 <-> v1alpha2`). Today the webhook's serving certificate **only** comes
from cert-manager: a `Certificate`/`Issuer` provisions the `…-tls` Secret, and the
`cert-manager.io/inject-ca-from` annotation lets cainjector populate the CRD's
`spec.conversion.webhook.clientConfig.caBundle`.

This makes cert-manager (with `cainjector`) a hard prerequisite. With the current
`certManager.enabled: false`, **no** certificate is produced at all — the webhook
pod crash-loops (cannot load TLS files) and conversion never works.

We want the chart to support **two** mutually exclusive certificate providers:

1. **cert-manager** — when it is already installed in the cluster (not installed by
   this chart). Continuous rotation + CA injection. Unchanged behavior.
2. **Self-signed** — the chart generates a long-lived (10+ year) self-signed cert
   so the webhook works with **no external dependency**.

## Non-goals

- Installing cert-manager from this chart (it is cluster-wide infrastructure; remains a
  documented prerequisite for the `certManager` provider).
- Automatic renewal of the self-signed cert. Sprig has no x509 expiry parser, so
  templates cannot read a cert's `notAfter`. Long duration + manual rotation instead.
- In-process cert rotation (cert-controller / rotator). Considered earlier and
  explicitly deferred in favor of a long-lived self-signed cert.
- Changing the webhook's startup cert-loading behavior in `cmd/webhook/main.go`
  (self-signed certs are stable; the cert-manager-rotation hot-reload is a separate,
  pre-existing item).

## Decisions (locked)

- **Deployment target:** primarily Helm CLI (`install`/`upgrade`), not GitOps. This
  makes Helm template crypto functions viable.
- **Generation mechanism:** Helm template functions (`genCA` + `genSignedCert`),
  not a pre-install Job. No extra image, no extra RBAC.
- **Values API:** explicit `conversionWebhook.tls.provider` enum
  (`certManager | selfSigned`). Chosen over reusing a boolean for readability and
  future extensibility (room for a later `provided` / bring-your-own-Secret mode).
  This is a **breaking values change** → chart version bump.
- **Renewal:** dropped. Long duration (default 3650 days) + reuse; rotation = delete
  the Secret and `helm upgrade`.

## Values API

```yaml
conversionWebhook:
  enabled: true
  # image, replicas, port, resources unchanged
  tls:
    provider: certManager        # certManager | selfSigned  (default preserves current behavior)
    certManager:
      issuerRef: {}              # moved from conversionWebhook.certManager.issuerRef
    selfSigned:
      durationDays: 3650         # 10 years; cert generated once and reused across upgrades
```

- Exactly one provider is active. `selfSigned.*` is ignored under `certManager` and
  vice-versa.
- Default `provider: certManager` keeps today's behavior for existing users (modulo the
  key relocation, see Migration).
- A template `fail` rejects any `provider` not in `{certManager, selfSigned}`.

### Migration (breaking change)

| Old (`0.2.0`)                              | New (`0.3.0`)                                   |
| ------------------------------------------ | ----------------------------------------------- |
| `conversionWebhook.certManager.enabled: true`  | `conversionWebhook.tls.provider: certManager`   |
| `conversionWebhook.certManager.enabled: false` | `conversionWebhook.tls.provider: selfSigned`    |
| `conversionWebhook.certManager.issuerRef`      | `conversionWebhook.tls.certManager.issuerRef`   |

Chart version `0.2.0 → 0.3.0`. Document in README + NOTES.txt.

## Architecture

### Certificate generation & reuse (self-signed)

A `_helpers.tpl` helper produces the cert material and **memoizes it on the root
context** so multiple templates in one render share the *same* CA:

1. `lookup` the existing `…-tls` Secret in the release namespace.
2. If it exists with `ca.crt` / `tls.crt` / `tls.key`, **reuse** those values.
3. Otherwise generate: `genCA(<name>-ca, durationDays)` then
   `genSignedCert(<name>, nil, <SANs>, durationDays, ca)`.
   SANs: `<name>.<ns>.svc` and `<name>.<ns>.svc.cluster.local` (match the existing
   cert-manager `Certificate` DNS names).
4. Store the result dict (`ca`, `cert`, `key`) on the root context under a private
   key (e.g. `_vitistackWebhookCerts`) so the Secret template and the CRD template
   read identical material. (`genSignedCert` is non-deterministic; without
   memoization the two resources would carry mismatched CAs.)

**Why `lookup`-reuse is required for correctness (not just churn avoidance):** the
webhook loads its TLS cert once at startup (`cmd/webhook/main.go`), and a content
change to the mounted Secret does not restart the pods. If an upgrade regenerated the
cert, the CRD `caBundle` would update while the running pods kept serving the old cert
→ the API server would reject the webhook → conversion breaks. Reuse keeps the cert
stable across upgrades so pods and `caBundle` stay in agreement.

**`checksum/tls` pod annotation:** the Deployment's pod template gets a
`checksum/tls` annotation derived from the cert material so that when the cert *does*
legitimately change (first install, or manual rotation after deleting the Secret), the
pods roll and pick it up.

### Resource matrix by provider

| Resource                              | `certManager`                                  | `selfSigned`                                   |
| ------------------------------------- | ---------------------------------------------- | ---------------------------------------------- |
| `Issuer` (self-signed fallback)       | rendered when no `issuerRef`                   | not rendered                                   |
| `Certificate` (cert-manager)          | rendered                                       | not rendered                                   |
| `…-tls` Secret                        | created by cert-manager                        | rendered by chart (new template)               |
| CRD `inject-ca-from` annotation       | present                                        | absent                                         |
| CRD inline `clientConfig.caBundle`    | absent (cainjector fills it)                   | present (generated CA, base64)                 |
| Deployment                            | unchanged + `checksum/tls`                     | unchanged + `checksum/tls`                     |

## Components / file changes

- `values.yaml` — replace `certManager` block with `tls.{provider,certManager,selfSigned}`.
- `Chart.yaml` — bump `version` to `0.3.0`.
- `templates/_helpers.tpl` — add the memoized self-signed cert helper + a provider
  validation helper (`fail` on unknown provider).
- `templates/webhook-selfsigned-secret.yaml` — **new**; rendered only for
  `provider == selfSigned`; writes `tls.crt` / `tls.key` / `ca.crt`.
- `templates/webhook-certificate.yaml` — guard switches to `provider == certManager`.
- `templates/vitistack.io_networknamespaces.yaml` — annotation vs inline `caBundle`
  conditioned on provider.
- `templates/webhook-deployment.yaml` — add `checksum/tls` pod annotation.
- `templates/NOTES.txt` — provider-aware messaging.
- `README.md` — "TLS / certificate provider" section (both modes + migration note).

## Error handling

- Unknown `tls.provider` → `fail` with a clear message at template time.
- `conversionWebhook.enabled: true` but no provider can produce a cert → `fail`
  (defensive; the enum makes this unreachable in practice).
- First install ordering: Helm creates the Secret (kind-sorted) before the Deployment,
  and the self-signed cert is materialized at render time, so the Secret already holds
  a valid cert when pods start.

## Testing

- `helm template` assertions per provider:
  - `certManager`: `Certificate` present, CRD has `inject-ca-from`, no inline `caBundle`,
    no chart-managed Secret.
  - `selfSigned`: chart Secret present with all three keys, CRD has inline non-empty
    `caBundle`, no `Certificate`/annotation.
- Verify the inline `caBundle` is the CA that signed the serving cert (decode + match).
- Extend/parallel `common/hack/test-conversion-webhook.sh` to cover the self-signed
  path end-to-end (apply, convert a `NetworkNamespace`, assert success without
  cert-manager present).

## Open risks

- `lookup` returns empty during `helm template`/`--dry-run`, so dry-run output shows a
  freshly generated cert each time. This is cosmetic for the Helm-CLI target (real
  `install`/`upgrade` use `lookup`); it would be a problem under GitOps, which is out
  of scope per the locked deployment-target decision.
