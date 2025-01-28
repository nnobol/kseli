#!/bin/bash

# Set Env Variables
export SECRET_KEY="super-secure-secret-key"
export API_KEY="super-secure-api-key"

# Build Frontend
cd client

if [ -z "$API_KEY" ]; then
    echo "Error: API_KEY environment variable is not set."
    exit 1
fi

VITE_API_KEY=$API_KEY npm run build

# Verify Build Success
if [ $? -ne 0 ]; then
    echo "Frontend build failed. Aborting deployment."
    exit 1
fi

# Swap Directories Atomically
cd ../builds
if [ -d "client-new" ]; then
    echo "Build complete. Swapping client directories."
    mv client client-old || true
    mv client-new client
    rm -rf client-old
    echo "Deployment successful."
else
    echo "Build directory client-new not found. Aborting."
    exit 1
fi

# Run Server
cd ..

cd server

go run cmd/main.go
