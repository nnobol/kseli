#!/bin/bash

# Build Frontend
cd client
npm run build

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
