#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

OUTPUT_DIR="${PROJECT_ROOT}/build/keys"
KEY_SIZE=2048

usage() {
  cat <<'EOF'
Usage:
  ./scripts/gen-keys.sh [options]

Options:
  --out <dir>      Output directory (default: ./keys)
  --size <bits>    Key size in bits (default: 2048)
  -h, --help       Show help

Examples:
  ./scripts/gen-keys.sh
  ./scripts/gen-keys.sh --out ./secrets
  ./scripts/gen-keys.sh --size 4096
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --out)
        shift
        OUTPUT_DIR="$1"
        ;;
      --size)
        shift
        KEY_SIZE="$1"
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        echo "Unknown option: $1"
        echo
        usage
        exit 1
        ;;
    esac
    shift
  done
}

main() {
  parse_args "$@"

  mkdir -p "${OUTPUT_DIR}"

  local private_key="${OUTPUT_DIR}/private.pem"
  local public_key="${OUTPUT_DIR}/public.pem"

  if [[ -f "${private_key}" ]] || [[ -f "${public_key}" ]]; then
    echo "Keys already exist in ${OUTPUT_DIR}:"
    [[ -f "${private_key}" ]] && echo "  ${private_key}"
    [[ -f "${public_key}" ]] && echo "  ${public_key}"
    echo
    read -r -p "Overwrite? [y/N] " confirm
    if [[ "${confirm}" != "y" && "${confirm}" != "Y" ]]; then
      echo "Aborted."
      exit 0
    fi
  fi

  echo "==> Generating RSA-${KEY_SIZE} private key"
  openssl genrsa -out "${private_key}" "${KEY_SIZE}"
  chmod 600 "${private_key}"

  echo "==> Extracting public key"
  openssl rsa -in "${private_key}" -pubout -out "${public_key}"
  chmod 644 "${public_key}"

  echo
  echo "Keys generated:"
  echo "  Private : ${private_key}"
  echo "  Public  : ${public_key}"
  echo
  echo "Add to your config:"
  echo "  auth:"
  echo "    private_key_path: \"${private_key}\""
}

main "$@"