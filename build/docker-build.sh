#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

IMAGE_NAME="file-server"
DOCKERFILE_PATH="${SCRIPT_DIR}/Dockerfile"
BUILD_SCRIPT="${PROJECT_ROOT}/scripts/build-linux.sh"
BIN_DIR="${SCRIPT_DIR}/bin"

TARGET_ARCH=""
IMAGE_TAG=""

usage() {
  cat <<'EOF'
Usage:
  ./build/docker-build.sh [options]

Options:
  --amd64              Build Docker image with linux/amd64 binary
  --arm64              Build Docker image with linux/arm64 binary
  --tag <tag>          Docker image tag (default: file-server:<arch>)
  --no-binary-build    Skip Go binary build and use existing binary
  -h, --help           Show help

Examples:
  ./build/docker-build.sh --amd64
  ./build/docker-build.sh --arm64
  ./build/docker-build.sh --arm64 --tag my-file-server:latest
  ./build/docker-build.sh --amd64 --no-binary-build
EOF
}

SKIP_BINARY_BUILD=false

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --amd64)
        if [[ -n "${TARGET_ARCH}" ]]; then
          echo "Only one architecture can be selected."
          exit 1
        fi
        TARGET_ARCH="amd64"
        ;;
      --arm64)
        if [[ -n "${TARGET_ARCH}" ]]; then
          echo "Only one architecture can be selected."
          exit 1
        fi
        TARGET_ARCH="arm64"
        ;;
      --tag)
        shift
        if [[ $# -eq 0 ]]; then
          echo "--tag requires a value"
          exit 1
        fi
        IMAGE_TAG="$1"
        ;;
      --no-binary-build)
        SKIP_BINARY_BUILD=true
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

  if [[ -z "${TARGET_ARCH}" ]]; then
    echo "You must select one architecture: --amd64 or --arm64"
    echo
    usage
    exit 1
  fi

  if [[ -z "${IMAGE_TAG}" ]]; then
    IMAGE_TAG="${IMAGE_NAME}:${TARGET_ARCH}"
  fi
}

build_binary() {
  echo "==> Building Linux binary for ${TARGET_ARCH}"
  "${BUILD_SCRIPT}" "--${TARGET_ARCH}"
}

verify_files() {
  local binary_path="${BIN_DIR}/fileServer-linux-${TARGET_ARCH}"
  local config_path="${SCRIPT_DIR}/config/config.yaml"

  if [[ ! -f "${binary_path}" ]]; then
    echo "Binary not found: ${binary_path}"
    exit 1
  fi

  if [[ ! -f "${config_path}" ]]; then
    echo "Config file not found: ${config_path}"
    exit 1
  fi
}

build_image() {
  local binary_name="fileServer-linux-${TARGET_ARCH}"

  echo "==> Building Docker image ${IMAGE_TAG}"
  docker build \
    -f "${DOCKERFILE_PATH}" \
    --build-arg BINARY_NAME="${binary_name}" \
    -t "${IMAGE_TAG}" \
    "${PROJECT_ROOT}"

  echo "Built Docker image: ${IMAGE_TAG}"
}

main() {
  parse_args "$@"

  if [[ "${SKIP_BINARY_BUILD}" != true ]]; then
    build_binary
  fi

  verify_files
  build_image

  echo
  echo "Done."
}

main "$@"