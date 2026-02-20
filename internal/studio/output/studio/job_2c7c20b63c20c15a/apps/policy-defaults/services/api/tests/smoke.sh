#!/usr/bin/env sh
set -eu
BASE_URL="${BASE_URL:-http://localhost:8090}"
curl -fsS "$BASE_URL/health" >/dev/null
curl -fsS "$BASE_URL/v1/tools" >/dev/null
echo "smoke-ok"
