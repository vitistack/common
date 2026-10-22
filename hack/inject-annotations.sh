#!/usr/bin/env bash
set -euo pipefail

# Wires the chart's common annotations (vitistack-crds.annotations: crds.keep and
# .Values.annotations) into every CRD *Helm template* after controller-gen output
# has been copied into the chart. Without this, crds.keep has no effect.
#
# Like inject-conversion.sh, it only touches the chart copy and runs on every
# `make gen-manifests`.

TEMPLATES_DIR=${1:-charts/vitistack-crds/templates}

for FILE in "$TEMPLATES_DIR"/vitistack.io_*.yaml; do
  if grep -q 'vitistack-crds.annotations' "$FILE"; then
    continue
  fi
  awk '
    { print }
    done==0 && /controller-gen\.kubebuilder\.io\/version:/ {
      print "    {{- with (include \"vitistack-crds.annotations\" . | trim) }}{{ . | nindent 4 }}{{- end }}"
      done=1
    }
  ' "$FILE" >"$FILE.tmp" && mv "$FILE.tmp" "$FILE"
done

echo "inject-annotations: wired common annotations into $TEMPLATES_DIR"
