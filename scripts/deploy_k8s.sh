#!/bin/bash

# Deploy K8s manifests with dynamic version injection
# Usage: ./deploy_k8s.sh <version>

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

DIST_DIR=".k8s_dist"
SRC_DIR="k8s"

# Clean up previous dist dir
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

echo "Preparing deployment for version $VERSION..."

# Copy manifests
cp -r "$SRC_DIR/"* "$DIST_DIR/"

# Copy .env from root (or k8s/ if root missing)
if [ -f ".env" ]; then
    cp ".env" "$DIST_DIR/.env"
elif [ -f "$SRC_DIR/.env" ]; then
    cp "$SRC_DIR/.env" "$DIST_DIR/.env"
else
    echo "Error: .env file not found in root or $SRC_DIR/"
    exit 1
fi

# Inject version into kustomization.yaml
# Replaces 'newTag: ...' with 'newTag: <VERSION>'
sed -i '' "s/newTag: .*/newTag: $VERSION/" "$DIST_DIR/kustomization.yaml"

echo "Applying manifests..."
kubectl apply -k "$DIST_DIR"

# Clean up
rm -rf "$DIST_DIR"
echo "Deployment applied successfully."
