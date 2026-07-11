#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

bash scripts/quality-check.sh
bash scripts/test-coverage.sh
python3 test_runtime_assets.py
