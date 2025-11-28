#!/bin/bash

set -e

echo "Building Samba Manager..."

# Build frontend
echo "Building frontend..."
cd frontend
pnpm install
pnpm run build

# Build backend with embedded frontend
echo "Building backend with embedded frontend..."
cd ..
go build -o samba-manager

echo "Build complete!"
echo "Binary location: ./samba-manager"
