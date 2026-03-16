#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/build.sh [target] [options]

Targets:
  linux         Build Linux binaries
  macos         Build macOS binaries
  all           Build Linux and macOS binaries

Options:
  --amd64       Build amd64 only
  --arm64       Build arm64 only
  --all         Build all architectures for selected target
  -h, --help    Show help

Examples:
  ./scripts/build.sh linux
  ./scripts/build.sh linux --arm64
  ./scripts/build.sh macos --amd64
  ./scripts/build.sh all
EOF
}

main() {
  if [[ $# -eq 0 ]]; then
    usage
    exit 0
  fi

  local target="$1"
  shift || true

  case "${target}" in
    linux)
      "$(dirname "$0")/build-linux.sh" "$@"
      ;;
    macos)
      "$(dirname "$0")/build-macos.sh" "$@"
      ;;
    all)
      "$(dirname "$0")/build-linux.sh" "$@"
      "$(dirname "$0")/build-macos.sh" "$@"
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Unknown target: ${target}"
      echo
      usage
      exit 1
      ;;
  esac
}

main "$@"