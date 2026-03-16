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
  ./scripts/build-macos.sh [options]

Options:
  --amd64        Build macOS amd64
  --arm64        Build macOS arm64
  --all          Build all macOS targets
  -h, --help     Show help

Examples:
  ./scripts/build-macos.sh
  ./scripts/build-macos.sh --all
  ./scripts/build-macos.sh --arm64
  ./scripts/build-macos.sh --amd64
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
  local binary_name="${APP_NAME}-darwin-${goarch}"
  local output_path="${OUTPUT_DIR}/${binary_name}"

  mkdir -p "${OUTPUT_DIR}"

  echo "==> Building darwin/${goarch}"

  GOOS=darwin GOARCH="${goarch}" CGO_ENABLED=0 \
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