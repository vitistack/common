#!/usr/bin/env bash
set -euo pipefail

DIR=${1:-crds}

# Remove integer formats that kubectl warns about
# Only removes lines exactly defining int32/int64 formats, preserves others like date-time.
shopt -s nullglob
for f in "$DIR"/*.yaml; do
  if command -v gsed >/dev/null 2>&1; then SED=gsed; else SED=sed; fi
  case "${OSTYPE:-}" in
    darwin*) INPLACE=(-i '') ;;
    *)       INPLACE=(-i) ;;
  esac
  $SED "${INPLACE[@]}" -E '/^[[:space:]]*format:[[:space:]]*"?int(32|64)"?[[:space:]]*$/d' "$f"
  echo "Sanitized integer formats in $f"
done
