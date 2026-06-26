#!/usr/bin/env bash
set -euo pipefail

# Exercises a locally-running NetworkNamespace conversion webhook by POSTing
# ConversionReview requests to /convert (both directions) plus the health probes.
#
# Pair it with the "Debug CRD Conversion Webhook" VS Code launch config:
#   1. make dev-certs           # once, to create .dev-certs/tls.{crt,key}
#   2. F5 -> "Debug CRD Conversion Webhook", set breakpoints in pkg/conversion
#   3. hack/test-conversion-webhook.sh
#
# Config (env or args):
#   WEBHOOK_URL  base URL of the webhook        (default https://localhost:9443)
#   first arg    overrides WEBHOOK_URL
#
# The server uses a self-signed dev cert, so curl runs with -k (insecure).

URL=${1:-${WEBHOOK_URL:-https://localhost:9443}}

pretty() {
  if command -v jq >/dev/null 2>&1; then jq .
  elif command -v python3 >/dev/null 2>&1; then python3 -m json.tool
  else cat
  fi
}

# send <title> <desiredAPIVersion> <object-json>
send() {
  local title=$1 desired=$2 object=$3 resp
  echo "=== ${title} ==="
  resp=$(curl -sk -X POST -H "Content-Type: application/json" "${URL}/convert" --data @- <<EOF
{
  "apiVersion": "apiextensions.k8s.io/v1",
  "kind": "ConversionReview",
  "request": {
    "uid": "test-$(echo "$title" | tr ' A-Z' '-a-z')",
    "desiredAPIVersion": "${desired}",
    "objects": [ ${object} ]
  }
}
EOF
  )
  echo "$resp" | pretty
  if echo "$resp" | grep -q '"status":"Success"'; then
    echo "  -> OK"
  else
    echo "  -> FAILED (no Success status)" >&2
    return 1
  fi
  echo ""
}

echo "Target: ${URL}"
echo ""

echo "=== health probes ==="
echo -n "  /healthz: "; curl -sk "${URL}/healthz"; echo ""
echo -n "  /readyz:  "; curl -sk "${URL}/readyz"; echo ""
echo ""

send "v1alpha1 to v1alpha2" "vitistack.io/v1alpha2" '{
  "apiVersion": "vitistack.io/v1alpha1",
  "kind": "NetworkNamespace",
  "metadata": { "name": "demo-ns", "namespace": "default" },
  "spec": { "datacenterIdentifier": "no-west-az1", "supervisorIdentifier": "my-namespace" }
}'

send "v1alpha2 to v1alpha1" "vitistack.io/v1alpha1" '{
  "apiVersion": "vitistack.io/v1alpha2",
  "kind": "NetworkNamespace",
  "metadata": { "name": "demo-ns", "namespace": "default" },
  "spec": {
    "datacenterIdentifier": "no-west-az1",
    "supervisorIdentifier": "my-namespace",
    "networkProvisioning": { "provider": "nam" }
  }
}'

echo "All conversion checks passed."
