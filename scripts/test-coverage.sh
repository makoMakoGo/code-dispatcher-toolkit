#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_DIR="$ROOT/code-dispatcher"
GO_TOOLCHAIN="${GOTOOLCHAIN:-go1.21.13+auto}"
threshold="${CODE_DISPATCHER_COVERAGE_MIN:-89}"
profile="${CODE_DISPATCHER_COVERAGE_PROFILE:-coverage.out}"

cd "$GO_DIR"

GOTOOLCHAIN="$GO_TOOLCHAIN" go test -race -coverprofile="$profile" ./...

coverage="$(GOTOOLCHAIN="$GO_TOOLCHAIN" go tool cover -func="$profile" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')"
if [[ -z "$coverage" ]]; then
  echo "failed to calculate coverage" >&2
  exit 1
fi

awk -v coverage="$coverage" -v threshold="$threshold" '
  BEGIN {
    if (coverage + 0 < threshold + 0) {
      printf("coverage %.1f%% is below required %.1f%%\n", coverage, threshold) > "/dev/stderr"
      exit 1
    }
    printf("coverage %.1f%% meets required %.1f%%\n", coverage, threshold)
  }
'
