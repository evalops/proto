#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

collect_workflows() {
  if [[ "${CHECK_ALL_WORKFLOWS:-false}" == "true" ]]; then
    find .github/workflows -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \) | sort
    return
  fi

  if [[ -n "${WORKFLOW_FILES:-}" ]]; then
    tr ' ' '\n' <<<"${WORKFLOW_FILES}" | sed '/^$/d' | sort -u
    return
  fi

  local diff_range=""
  if [[ "${GITHUB_EVENT_NAME:-}" == "pull_request" && -n "${GITHUB_BASE_REF:-}" ]]; then
    git fetch origin "${GITHUB_BASE_REF}" --depth=1 >/dev/null 2>&1 || true
    diff_range="origin/${GITHUB_BASE_REF}...HEAD"
  elif [[ -n "${GITHUB_SHA:-}" ]] && git rev-parse --verify "${GITHUB_SHA}^{commit}" >/dev/null 2>&1; then
    if git rev-parse --verify "${GITHUB_SHA}~1^{commit}" >/dev/null 2>&1; then
      diff_range="${GITHUB_SHA}~1...${GITHUB_SHA}"
    fi
  elif git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
    diff_range="HEAD~1...HEAD"
  fi

  if [[ -n "${diff_range}" ]]; then
    git diff --name-only "${diff_range}" -- .github/workflows | grep -E '\.ya?ml$' || true
  else
    find .github/workflows -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \) | sort
  fi
}

mapfile -t workflow_files < <(collect_workflows)
if (( ${#workflow_files[@]} == 0 )); then
  echo "No workflow files to validate."
  exit 0
fi

failures=()
warnings=()

for workflow in "${workflow_files[@]}"; do
  if [[ ! -f "${workflow}" ]]; then
    continue
  fi

  if ! grep -qE '^[[:space:]]*permissions:' "${workflow}"; then
    failures+=("${workflow}: missing explicit permissions block")
  fi

  while IFS= read -r entry; do
    line_no="${entry%%:*}"
    line_text="${entry#*:}"
    uses_ref="$(sed -E 's/^[[:space:]]*uses:[[:space:]]*([^[:space:]#]+).*/\1/' <<<"${line_text}")"

    if [[ -z "${uses_ref}" || "${uses_ref}" == "${line_text}" ]]; then
      continue
    fi
    if [[ "${uses_ref}" == ./* || "${uses_ref}" == docker://* ]]; then
      continue
    fi
    if [[ "${uses_ref}" != *@* ]]; then
      failures+=("${workflow}:${line_no} uses '${uses_ref}' without @ref")
      continue
    fi

    ref="${uses_ref##*@}"
    if [[ "${ref}" =~ ^[0-9a-f]{40}$ ]]; then
      continue
    fi
    if [[ "${ref}" =~ ^(main|master|HEAD|latest)$ ]]; then
      failures+=("${workflow}:${line_no} uses floating ref '${uses_ref}'")
      continue
    fi

    if [[ "${ENFORCE_SHA_PINS:-false}" == "true" ]]; then
      failures+=("${workflow}:${line_no} uses '${uses_ref}' without full-length SHA pin")
    else
      warnings+=("${workflow}:${line_no} non-SHA pin '${uses_ref}' (allowed while ENFORCE_SHA_PINS=false)")
    fi
  done < <(grep -nE '^[[:space:]]*uses:[[:space:]]*[^#[:space:]].*$' "${workflow}" || true)
done

if (( ${#failures[@]} > 0 )); then
  echo "Workflow security baseline violations found:" >&2
  for violation in "${failures[@]}"; do
    echo "- ${violation}" >&2
  done
  exit 1
fi

if (( ${#warnings[@]} > 0 )); then
  echo "Workflow security baseline warnings:" >&2
  for warning in "${warnings[@]}"; do
    echo "- ${warning}" >&2
  done
fi

echo "Workflow security baseline checks passed for ${#workflow_files[@]} workflow file(s)."
