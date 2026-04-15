#!/usr/bin/env bash
set -euo pipefail
CGO_ENABLED=0 go build -v -o gotimekpr ./cmd/gotimekpr