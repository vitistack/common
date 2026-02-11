#!/usr/bin/env bash
#
# Clear GitHub Actions caches for vitistack repositories.
#
# Usage:
#   ./clear-gh-cache.sh                     # Clear caches for ALL vitistack repos
#   ./clear-gh-cache.sh common              # Clear caches for vitistack/common
#   ./clear-gh-cache.sh common kea-operator # Clear caches for multiple repos
#   ./clear-gh-cache.sh --dry-run common    # Show what would be deleted
#
# Requirements:
#   - gh CLI (https://cli.github.com/) authenticated with appropriate permissions
#

set -euo pipefail

ORG="vitistack"
DRY_RUN=false

usage() {
  cat <<EOF
Usage: $(basename "$0") [--dry-run] [repo1 repo2 ...]

Clear GitHub Actions caches for vitistack repositories.

Options:
  --dry-run    Show what would be deleted without actually deleting
  --help       Show this help message

Arguments:
  repo1 repo2  One or more repository names (without org prefix).
               If none specified, all repos in ${ORG} will be processed.

Examples:
  $(basename "$0")                          # All repos
  $(basename "$0") common                   # Single repo
  $(basename "$0") common kea-operator      # Multiple repos
  $(basename "$0") --dry-run common         # Dry run
EOF
  exit 0
}

# Parse flags
REPOS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --help|-h)
      usage
      ;;
    *)
      REPOS+=("$1")
      shift
      ;;
  esac
done

# Check gh CLI is available and authenticated
if ! command -v gh &>/dev/null; then
  echo "Error: gh CLI is not installed. Install from https://cli.github.com/"
  exit 1
fi

if ! gh auth status &>/dev/null; then
  echo "Error: gh CLI is not authenticated. Run 'gh auth login' first."
  exit 1
fi

# If no repos specified, list all repos in the org
if [[ ${#REPOS[@]} -eq 0 ]]; then
  echo "Fetching all repositories for ${ORG}..."
  mapfile -t REPOS < <(gh repo list "${ORG}" --limit 200 --json name --jq '.[].name' | sort)
  echo "Found ${#REPOS[@]} repositories"
fi

TOTAL_DELETED=0
TOTAL_ERRORS=0

for REPO in "${REPOS[@]}"; do
  FULL_REPO="${ORG}/${REPO}"
  echo ""
  echo "=== ${FULL_REPO} ==="

  # List all caches
  CACHES=$(gh cache list --repo "${FULL_REPO}" --json id,key,sizeInBytes --limit 100 2>/dev/null || echo "[]")
  COUNT=$(echo "${CACHES}" | jq 'length')

  if [[ "${COUNT}" -eq 0 ]]; then
    echo "  No caches found"
    continue
  fi

  # Calculate total size
  TOTAL_SIZE=$(echo "${CACHES}" | jq '[.[].sizeInBytes] | add')
  TOTAL_SIZE_MB=$(echo "scale=1; ${TOTAL_SIZE} / 1048576" | bc 2>/dev/null || echo "?")

  echo "  Found ${COUNT} cache(s) (${TOTAL_SIZE_MB} MB)"

  # Delete each cache
  echo "${CACHES}" | jq -r '.[] | "\(.id)\t\(.key)"' | while IFS=$'\t' read -r ID KEY; do
    if [[ "${DRY_RUN}" == "true" ]]; then
      echo "  [dry-run] Would delete: ${KEY} (id: ${ID})"
    else
      if gh cache delete "${ID}" --repo "${FULL_REPO}" 2>/dev/null; then
        echo "  Deleted: ${KEY}"
      else
        echo "  Failed to delete: ${KEY} (id: ${ID})"
      fi
    fi
  done

  TOTAL_DELETED=$((TOTAL_DELETED + COUNT))
done

echo ""
if [[ "${DRY_RUN}" == "true" ]]; then
  echo "Dry run complete. ${TOTAL_DELETED} cache(s) would be deleted."
else
  echo "Done. Processed ${TOTAL_DELETED} cache(s)."
fi
