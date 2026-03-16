#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

APP_NAME="fileServer"
MAIN_PKG="${PROJECT_ROOT}/cmd"
OUTPUT_DIR="${PROJECT_ROOT}/build/bin"

BUILD_AMD64=false
BUILD_ARM64=false

usage() {
  cat <<'EOF'
Usage:
  ./scripts/build-linux.sh [options]

Options:
  --amd64        Build Linux amd64
  --arm64        Build Linux arm64
  --all          Build all Linux targets
  -h, --help     Show help

Examples:
  ./scripts/build-linux.sh
  ./scripts/build-linux.sh --all
  ./scripts/build-linux.sh --amd64
  ./scripts/build-linux.sh --arm64
  ./scripts/build-linux.sh --amd64 --arm64
EOF
}

parse_args() {
  if [[ $# -eq 0 ]]; then
    BUILD_AMD64=true
    BUILD_ARM64=true
    return
  fi

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --amd64)
        BUILD_AMD64=true
        ;;
      --arm64)
        BUILD_ARM64=true
        ;;
      --all)
        BUILD_AMD64=true
        BUILD_ARM64=true
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

build_target() {
  local goarch="$1"
  local binary_name="${APP_NAME}-linux-${goarch}"
  local output_path="${OUTPUT_DIR}/${binary_name}"

  mkdir -p "${OUTPUT_DIR}"

  echo "==> Building linux/${goarch}"

  GOOS=linux GOARCH="${goarch}" CGO_ENABLED=0 \
    go build -o "${output_path}" "${MAIN_PKG}"

  echo "Built: ${output_path}"
}

main() {
  parse_args "$@"

  if [[ "${BUILD_AMD64}" == true ]]; then
    build_target "amd64"
  fi

  if [[ "${BUILD_ARM64}" == true ]]; then
    build_target "arm64"
  fi

  echo
  echo "Done."
}

main "$@"