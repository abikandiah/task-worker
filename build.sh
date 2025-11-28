#!/bin/bash

# --- Configuration ---
IMAGE_NAME="task-worker"
DOCKERFILE_PATH="./Dockerfile"
# ---------------------

# Exit immediately if a command exits with a non-zero status.
set -e

echo "--- Starting Docker Image Build Process ---"

# Get the latest Git tag
# Use '|| true' to let Bash continue even if 'git describe' exits with an error (e.g., no tags).
GIT_VERSION_TAG=$(git describe --tags --abbrev=0 2>/dev/null || true)

if [ -z "$GIT_VERSION_TAG" ]; then
    echo "WARNING: Could not determine a valid Git tag. Using 'latest'."
    GIT_VERSION_TAG="latest"
fi

# Define the full image tag
FULL_IMAGE_TAG="${IMAGE_NAME}:${GIT_VERSION_TAG}"

echo "Git Version Tag determined: ${GIT_VERSION_TAG}"
echo "Building image with full tag: ${FULL_IMAGE_TAG}"

# Execute the Docker build command
# The '.' is the build context.
sudo docker build -f "${DOCKERFILE_PATH}" -t "${FULL_IMAGE_TAG}" .

# --- Script End ---

if [ $? -eq 0 ]; then
    echo "--- Build SUCCESSFUL ---"
    echo "Image created: ${FULL_IMAGE_TAG}"
else
    echo "--- Build FAILED ---"
    exit 1
fi
