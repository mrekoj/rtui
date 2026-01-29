#!/usr/bin/env bash
set -euo pipefail

scripts/phase1_tests.sh
scripts/phase2_tests.sh
scripts/phase3_tests.sh
scripts/smoke.sh

GOFLAGS="-count=1" go test ./...
