#!/usr/bin/env bash
set -euo pipefail

project_id=""
region=""
repository=""
scope=""
output_path=""
access_token=""

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --project-id)
      project_id="$2"
      shift 2
      ;;
    --region)
      region="$2"
      shift 2
      ;;
    --repository)
      repository="$2"
      shift 2
      ;;
    --scope)
      scope="$2"
      shift 2
      ;;
    --output)
      output_path="$2"
      shift 2
      ;;
    --access-token)
      access_token="$2"
      shift 2
      ;;
    *)
      echo "error: unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "${project_id}" || -z "${region}" || -z "${repository}" || -z "${scope}" || -z "${output_path}" ]]; then
  cat >&2 <<'EOF'
error: missing required arguments.
usage:
  write-npmrc.sh \
    --project-id <project-id> \
    --region <region> \
    --repository <repository> \
    --scope <scope> \
    --output <path> \
    [--access-token <token>]
EOF
  exit 1
fi

if [[ "${scope}" != @* ]]; then
  echo "error: scope must start with '@'." >&2
  exit 1
fi

if [[ -z "${access_token}" ]]; then
  access_token="$(gcloud auth print-access-token)"
fi

if [[ -z "${access_token}" ]]; then
  echo "error: access token was empty." >&2
  exit 1
fi

registry_host="${region}-npm.pkg.dev"
registry_path="${registry_host}/${project_id}/${repository}/"
registry_url="https://${registry_path}"

mkdir -p "$(dirname "${output_path}")"
umask 077
cat > "${output_path}" <<EOF
${scope}:registry=${registry_url}
//${registry_path}:_authToken=${access_token}
always-auth=true
EOF

echo "wrote ${output_path} for ${scope} -> ${registry_url}" >&2
