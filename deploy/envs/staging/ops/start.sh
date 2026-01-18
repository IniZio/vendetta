#!/bin/bash
# Start coordination server

set -e

cd "$(dirname "$0")/../.."
SERVER_ROOT="${PWD}"
STAGING_DIR="${SERVER_ROOT}/deploy/envs/staging"

# Use staging config
export VENDETTA_COORD_CONFIG="${STAGING_DIR}/config/coordination.yaml"

echo "=== Starting Nexus Coordination Server ==="
echo "Config: ${VENDETTA_COORD_CONFIG}"
echo "Server: http://localhost:3001"
echo ""
echo "To stop: Ctrl+C"
echo ""

cd "${SERVER_ROOT}"
nexus coordination start
