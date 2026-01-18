#!/bin/bash
# Workspace management for staging

set -e

SERVER="${SERVER:-http://localhost:3001}"

create() {
    local github_username="$1"
    local workspace_name="$2"
    
    if [[ -z "$github_username" || -z "$workspace_name" ]]; then
        echo "Usage: ./workspaces.sh create <github-username> <workspace-name>"
        echo ""
        echo "Example:"
        echo "  ./workspaces.sh create alice feature-x"
        exit 1
    fi
    
    echo "Creating workspace: $workspace_name (user: $github_username)"
    
    curl -s -X POST "$SERVER/api/v1/workspaces/create-from-repo" \
        -H "Content-Type: application/json" \
        -d "{
            \"github_username\": \"$github_username\",
            \"workspace_name\": \"$workspace_name\",
            \"repo\": {
                \"owner\": \"oursky\",
                \"name\": \"epson-eshop\",
                \"url\": \"git@github.com:oursky/epson-eshop.git\",
                \"branch\": \"main\",
                \"is_fork\": false
            },
            \"provider\": \"lxc\",
            \"image\": \"ubuntu:22.04\",
            \"services\": [
                {\"name\": \"web\", \"command\": \"bundle exec puma -p 5000\", \"port\": 5000}
            ]
        }" | jq '.'
}

list() {
    echo "Listing workspaces..."
    curl -s "$SERVER/api/v1/workspaces" | jq '.'
}

status() {
    local workspace_id="$1"
    
    if [[ -z "$workspace_id" ]]; then
        echo "Usage: ./workspaces.sh status <workspace-id>"
        exit 1
    fi
    
    echo "Workspace status: $workspace_id"
    curl -s "$SERVER/api/v1/workspaces/$workspace_id/status" | jq '.'
}

stop() {
    local workspace_id="$1"
    
    if [[ -z "$workspace_id" ]]; then
        echo "Usage: ./workspaces.sh stop <workspace-id>"
        exit 1
    fi
    
    echo "Stopping workspace: $workspace_id"
    curl -s -X POST "$SERVER/api/v1/workspaces/$workspace_id/stop" | jq '.'
}

delete() {
    local workspace_id="$1"
    
    if [[ -z "$workspace_id" ]]; then
        echo "Usage: ./workspaces.sh delete <workspace-id>"
        exit 1
    fi
    
    echo "WARNING: Deleting workspace: $workspace_id"
    read -p "Continue? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        curl -s -X DELETE "$SERVER/api/v1/workspaces/$workspace_id" | jq '.'
    fi
}

case "$1" in
    create)
        create "$2" "$3"
        ;;
    list)
        list
        ;;
    status)
        status "$2"
        ;;
    stop)
        stop "$2"
        ;;
    delete)
        delete "$2"
        ;;
    *)
        echo "Usage: ./workspaces.sh <command>"
        echo ""
        echo "Commands:"
        echo "  create <username> <name>      Create workspace"
        echo "  list                          List all workspaces"
        echo "  status <id>                   Get workspace status"
        echo "  stop <id>                     Stop workspace"
        echo "  delete <id>                   Delete workspace"
        exit 1
        ;;
esac
