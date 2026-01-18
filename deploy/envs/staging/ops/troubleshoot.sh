#!/bin/bash
# Troubleshooting for staging environment

SERVER="${SERVER:-http://localhost:3001}"

echo "=== Nexus Staging Troubleshooting ==="
echo ""

echo "[1/5] Check if server is running..."
if curl -s "$SERVER/health" > /dev/null 2>&1; then
    echo "✓ Server is UP"
else
    echo "✗ Server is DOWN"
    echo "  Start: ./ops/start.sh"
    exit 1
fi

echo ""
echo "[2/5] Check port availability..."
if lsof -i :3001 > /dev/null 2>&1; then
    echo "✓ Port 3001 in use"
else
    echo "✗ Port 3001 not in use (expected if server running)"
fi

echo ""
echo "[3/5] Check LXC availability..."
if command -v lxc &> /dev/null; then
    echo "✓ LXC installed"
    lxc version
else
    echo "✗ LXC not installed"
    echo "  Install: apt-get install lxc (Ubuntu) or brew install lxc (Mac)"
fi

echo ""
echo "[4/5] Check SSH key..."
if [[ -f ~/.ssh/id_ed25519.pub ]]; then
    echo "✓ SSH key exists"
    echo "  Pubkey: $(cat ~/.ssh/id_ed25519.pub | head -c 50)..."
else
    echo "✗ SSH key missing"
    echo "  Generate: ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519 -N ''"
fi

echo ""
echo "[5/5] List active workspaces..."
if curl -s "$SERVER/api/v1/workspaces" 2>/dev/null | jq -e '.workspaces' > /dev/null 2>&1; then
    count=$(curl -s "$SERVER/api/v1/workspaces" | jq '.workspaces | length')
    echo "✓ Found $count workspace(s)"
    curl -s "$SERVER/api/v1/workspaces" | jq '.workspaces[] | {id, name, status}' 2>/dev/null || echo "  (none)"
else
    echo "✗ Cannot list workspaces"
    echo "  Check server logs: ./ops/start.sh"
fi

echo ""
echo "=== Done ==="
