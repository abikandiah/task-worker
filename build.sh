#!/bin/bash

# --- Configuration ---
IMAGE_NAME="task-worker"
DOCKERFILE_PATH="./Dockerfile"
# ---------------------

# Exit immediately if a command exits with a non-zero status.
set -e

# Get the latest Git tag
# Use '|| true' to let Bash continue even if 'git describe' exits with an error (e.g., no tags).
GIT_VERSION_TAG=$(git describe --tags --abbrev=0 2>/dev/null || true)

if [ -z "$GIT_VERSION_TAG" ]; then
    echo "WARNING: Could not determine a valid Git tag. Using 'latest'."
    GIT_VERSION_TAG="latest"
fi
echo "Git Version Tag determined: ${GIT_VERSION_TAG}"

echo "=========================================="
echo "🚀 Starting Local CI Pipeline (Version: $GIT_VERSION_TAG)"
echo "=========================================="
echo ""

# --- Stage 1: Testing ---
echo "--- Stage 1: Running Tests ---"
echo ""

# --- Stage 2: Build and Copy UI ---
echo "--- Stage 2: Building and Copying UI ---"
build_and_sync_ui() {
	cd ../task-executor-ui
	npm install
	npm run build

	echo "Finished building UI"

	cd ../task-worker
	echo "Synchronizing UI assets to ./dist"
	rsync -av --delete "../task-executor-ui/dist/" "./dist"
}

if build_and_sync_ui; then
	echo "UI copied successfully"
else 
	echo "❌ UI copy failed"
	exit 1
fi
echo ""

# --- Stage 3: Build Server ---
echo "--- Stage 3: Building App ---"
if go build -o ./bin/server ./cmd/api; then
	echo "App built successfully"
else
	echo "❌ App build failed"
	exit 1
fi
echo ""

# --- Stage 4: Docker Build ---
echo "--- Stage 4: Building Docker Image ---"
echo ""

FULL_IMAGE_TAG="${IMAGE_NAME}:${GIT_VERSION_TAG}"
echo "Building image with full tag: ${FULL_IMAGE_TAG}"

# Execute the Docker build command
# The '.' is the build context.
if sudo docker build -f "${DOCKERFILE_PATH}" -t "${FULL_IMAGE_TAG}" .; then
    echo "✅ Docker Image Built Successfully: $FULL_IMAGE_TAG"
else
    echo "❌ Docker Build Failed. Review Dockerfile and logs."
    exit 1
fi
echo ""

echo "=========================================="
echo "Pipeline COMPLETE."
echo "To run the container: docker run -d -p 8080:8080 $FULL_IMAGE_TAG"
echo "=========================================="
echo ""
