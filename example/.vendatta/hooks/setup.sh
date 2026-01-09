#!/bin/bash
# Setup script for example project
echo "Setting up environment..."

# Install Node.js if not present
if ! command -v node &> /dev/null; then
    echo "Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
    apt-get install -y nodejs
fi

# Install docker-compose if not present
if ! command -v docker-compose &> /dev/null; then
    echo "Installing docker-compose..."
    apt-get update && apt-get install -y docker-compose
fi

# Start services in background
echo "Starting database..."
cd /workspace && docker-compose up -d &
DB_PID=$!

echo "Waiting for database..."
sleep 10

echo "Starting API server..."
cd /workspace/server && npm install && HOST=0.0.0.0 PORT=5000 npm run dev &
API_PID=$!

sleep 5

echo "Starting web client..."
cd /workspace/client && npm install && HOST=0.0.0.0 PORT=3000 npm run dev &
WEB_PID=$!

echo "Services starting... PIDs: DB($DB_PID), API($API_PID), WEB($WEB_PID)"
echo "Setup complete. Services will be available shortly."
