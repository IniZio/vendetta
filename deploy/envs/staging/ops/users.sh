#!/bin/bash
# User management for staging

set -e

SERVER="${SERVER:-http://localhost:3001}"

register() {
    local github_username="$1"
    local github_id="$2"
    local ssh_pubkey="$3"
    
    if [[ -z "$github_username" || -z "$github_id" || -z "$ssh_pubkey" ]]; then
        echo "Usage: ./users.sh register <github-username> <github-id> <ssh-pubkey>"
        echo ""
        echo "Example:"
        echo "  ./users.sh register alice 123456789 \"ssh-ed25519 AAAA...\""
        echo ""
        echo "Get SSH pubkey:"
        echo "  cat ~/.ssh/id_ed25519.pub"
        exit 1
    fi
    
    # Extract fingerprint
    ssh_fingerprint=$(echo "$ssh_pubkey" | ssh-keygen -l -f - 2>/dev/null | awk '{print $2}' || echo "SHA256:unknown")
    
    echo "Registering user: $github_username"
    
    curl -s -X POST "$SERVER/api/v1/users/register-github" \
        -H "Content-Type: application/json" \
        -d "{
            \"github_username\": \"$github_username\",
            \"github_id\": $github_id,
            \"ssh_pubkey\": \"$ssh_pubkey\",
            \"ssh_pubkey_fingerprint\": \"$ssh_fingerprint\"
        }" | jq '.'
}

case "$1" in
    register)
        register "$2" "$3" "$4"
        ;;
    *)
        echo "Usage: ./users.sh <command>"
        echo ""
        echo "Commands:"
        echo "  register <username> <id> <pubkey>  Register new user"
        exit 1
        ;;
esac
