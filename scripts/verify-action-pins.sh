#!/usr/bin/env bash
# Verify that every SHA-pinned GitHub Action reference in workflow files
# resolves to an actual commit. Catches hallucinated hashes from AI agents.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

workflow_dir=".github/workflows"
if [[ ! -d "${workflow_dir}" ]]; then
  echo "No .github/workflows directory found."
  exit 0
fi

failures=()
checked=0

for workflow in "${workflow_dir}"/*.yml "${workflow_dir}"/*.yaml; do
  [[ -f "${workflow}" ]] || continue

  while IFS= read -r entry; do
    line_no="${entry%%:*}"
    line_text="${entry#*:}"

    uses_ref="$(sed -E 's/^[[:space:]]*uses:[[:space:]]*([^[:space:]#]+).*/\1/' <<<"${line_text}")"
    [[ -n "${uses_ref}" && "${uses_ref}" != "${line_text}" ]] || continue
    [[ "${uses_ref}" != ./* && "${uses_ref}" != docker://* ]] || continue
    [[ "${uses_ref}" == *@* ]] || continue

    owner_repo="${uses_ref%%@*}"
    ref="${uses_ref##*@}"

    # Only check full SHA pins
    [[ "${ref}" =~ ^[0-9a-f]{40}$ ]] || continue

    checked=$((checked + 1))

    # Verify the commit exists via the GitHub API
    http_code=$(curl -s -o /dev/null -w '%{http_code}' \
      -H "Accept: application/vnd.github+json" \
      ${GITHUB_TOKEN:+-H "Authorization: Bearer ${GITHUB_TOKEN}"} \
      "https://api.github.com/repos/${owner_repo}/git/commits/${ref}" 2>/dev/null || echo "000")

    if [[ "${http_code}" != "200" ]]; then
      # Extract the version comment if present
      version_comment="$(sed -E 's/.*#[[:space:]]*(v[0-9][^ ]*).*/\1/' <<<"${line_text}")"
      failures+=("${workflow}:${line_no}: ${owner_repo}@${ref:0:12}... (${version_comment}) — HTTP ${http_code}, commit does not exist")
    fi
  done < <(grep -nE '^[[:space:]]*uses:[[:space:]]*[^#[:space:]].*$' "${workflow}" || true)
done

if (( ${#failures[@]} > 0 )); then
  echo "Action pin verification failed — hallucinated or invalid commit hashes:" >&2
  for f in "${failures[@]}"; do
    echo "  ✗ ${f}" >&2
  done
  echo "" >&2
  echo "Fix: look up the real commit hash with:" >&2
  echo "  gh api repos/OWNER/REPO/git/refs/tags/VERSION --jq '.object.sha'" >&2
  exit 1
fi

echo "Verified ${checked} SHA-pinned action reference(s) across $(find "${workflow_dir}" -name '*.yml' -o -name '*.yaml' | wc -l | tr -d ' ') workflow file(s)."
