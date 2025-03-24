#!/bin/bash

# Exit on error
set -e

echo "Building Windows executable using Docker..."

# Create output directory
mkdir -p ./releases/windows

# Build Docker image
docker build -t attendance-tracker-windows-builder -f Dockerfile.windows .

# Extract the built executable from the container
docker create --name temp-builder attendance-tracker-windows-builder
docker cp temp-builder:/output/attendance-tracker-windows-amd64.zip ./releases/windows/
docker rm temp-builder

# Unzip the package
cd ./releases/windows
unzip -o attendance-tracker-windows-amd64.zip
rm attendance-tracker-windows-amd64.zip  # Optional: remove the zip after extraction

echo "Build complete! Windows executable is available in ./releases/windows/" 
