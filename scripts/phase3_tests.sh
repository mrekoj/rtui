#!/usr/bin/env bash
set -euo pipefail

GOFLAGS="-count=1" go test ./internal/ui
