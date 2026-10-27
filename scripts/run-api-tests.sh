#!/usr/bin/env bash
set -e

# Ensure tmp directory exists
mkdir -p ./tmp

# Clean up database
rm -f ./data/gorm.db

# Download Postman collection
echo "Downloading Postman collection..."
# The spec moved to realworld-apps/realworld and the Postman collection was removed in 0e68ba1;
# pin the last commit that still has it, so the download does not break again
COLLECTION_SHA=58acb59604f32db6f887b349a7cdc636760e2337
curl -L -s -f "https://raw.githubusercontent.com/realworld-apps/realworld/${COLLECTION_SHA}/specs/api/legacy_Conduit.postman_collection.json" -o ./tmp/Conduit.postman_collection.json

# Build the application
echo "Building application..."
go build -o app hello.go

# Start the server
echo "Starting server..."
PORT=8080 JWT_SECRET="${JWT_SECRET:-$(head -c 48 /dev/urandom | base64)}" ./app &
SERVER_PID=$!

# Cleanup function to kill server on exit
cleanup() {
    echo "Stopping server..."
    kill $SERVER_PID
    rm -f app
}
trap cleanup EXIT

# Wait for server to be ready
echo "Waiting for server to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8080/api/ping > /dev/null; then
        echo "Server is up!"
        break
    fi
    sleep 1
done

# Run Newman
echo "Running API tests..."
# Check if newman is available
if ! command -v newman &> /dev/null; then
    echo "newman not found, trying npx..."
    npx newman run ./tmp/Conduit.postman_collection.json \
      --global-var "APIURL=http://localhost:8080/api" \
      --global-var "EMAIL=test@example.com" \
      --global-var "PASSWORD=password" \
      --global-var "USERNAME=testuser" \
      --delay-request 50
else
    newman run ./tmp/Conduit.postman_collection.json \
      --global-var "APIURL=http://localhost:8080/api" \
      --global-var "EMAIL=test@example.com" \
      --global-var "PASSWORD=password" \
      --global-var "USERNAME=testuser" \
      --delay-request 50
fi
