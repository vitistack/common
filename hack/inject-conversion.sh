#!/usr/bin/env bash
set -euo pipefail

# Injects the conversion-webhook wiring into the NetworkNamespace CRD *Helm
# template* after controller-gen output has been copied into the chart.
#
# controller-gen cannot emit spec.conversion (it doesn't know the Service
# name/namespace, which are Helm release values), and `make gen-manifests`
# overwrites the chart's copy of the CRD on every run — so this runs as part of
# that flow to re-apply the wiring each time. It only touches the chart copy;
# the raw crds/ output and crds.yaml (the kubectl-apply artifact) stay free of
# Helm syntax.
#
# Wiring added (both gated on .Values.conversionWebhook.enabled):
#   - metadata.annotations: cert-manager.io/inject-ca-from  (cainjector fills caBundle)
#   - spec.conversion: strategy Webhook -> the conversion-webhook Service

TEMPLATES_DIR=${1:-charts/vitistack-crds/templates}
FILE="$TEMPLATES_DIR/vitistack.io_networknamespaces.yaml"

if [[ ! -f "$FILE" ]]; then
  echo "inject-conversion: $FILE not found, skipping" >&2
  exit 0
fi

if grep -q 'cert-manager.io/inject-ca-from' "$FILE"; then
  echo "inject-conversion: wiring already present in $FILE, skipping"
  exit 0
fi

awk '
  { print }
  ann==0 && /controller-gen\.kubebuilder\.io\/version:/ {
    print "    {{- if .Values.conversionWebhook.enabled }}"
    print "    cert-manager.io/inject-ca-from: {{ .Release.Namespace }}/{{ include \"vitistack-crds.webhook.name\" . }}-cert"
    print "    {{- end }}"
    ann=1
  }
  spec==0 && /^spec:[[:space:]]*$/ {
    print "  {{- if .Values.conversionWebhook.enabled }}"
    print "  conversion:"
    print "    strategy: Webhook"
    print "    webhook:"
    print "      conversionReviewVersions:"
    print "        - v1"
    print "      clientConfig:"
    print "        service:"
    print "          namespace: {{ .Release.Namespace }}"
    print "          name: {{ include \"vitistack-crds.webhook.name\" . }}"
    print "          path: /convert"
    print "          port: 443"
    print "  {{- end }}"
    spec=1
  }
' "$FILE" >"$FILE.tmp" && mv "$FILE.tmp" "$FILE"

echo "inject-conversion: wired conversion webhook into $FILE"
