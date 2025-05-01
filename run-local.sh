#!/bin/bash

set -e

# Set Env Variables
export ENV="local"
export SECRET_KEY="super-secure-secret-key"
export API_KEY="super-secure-api-key"

echo "Building client..."
cd client
VITE_API_KEY=$API_KEY npm run build
cd ..

cd builds
if [ -d "client-new" ]; then
    echo "Build complete. Replacing old client build..."
    rm -rf client
    mv client-new client
    echo "Swapped to new client build."
else
    echo "Build directory client-new not found. Aborting."
    exit 1
fi
cd ..

echo "Starting Go server..."
cd server
go run cmd/server/main.go
