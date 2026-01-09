#!/bin/bash
# Setup script for example project
echo "Setting up environment..."

# Start services in background
echo "Starting database..."
docker-compose up -d postgres &

echo "Waiting for database..."
sleep 5

echo "Starting API server..."
cd server && npm install && npm run dev &
API_PID=$!

echo "Starting web client..."
cd client && npm install && npm run dev &
WEB_PID=$!

echo "Services started. PIDs: DB(background), API($API_PID), WEB($WEB_PID)"
echo "Setup complete."
