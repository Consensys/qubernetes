#!/usr/bin/env sh

# Helper for building dockerhub images from specified git tags.
# ./docker-build.sh v0.1.1 quorumengineering/qubernetes

# To see labels:
# docker image inspect --format='' $IMAGE_NAME

# For debugging all env vars
# printenv

IMAGE_NAME="quorumengineering/qubernetes"
VERSION_TAG="latest"

usage() {
  echo "usage: "
  echo "  docker-build.sh VERSION IMAGE_NAME"
  echo "  docker-build.sh v0.1.1 quorumengineering/qubernetes"
  echo
}

if [[ $# -lt 1 ]]; then
  echo " No version or image_name passed in, using defaults. "
  echo
  usage
  echo
fi

if [[ $# -gt 0 ]]; then
  VERSION_TAG=$1
fi

if [[ $# -gt 1 ]]; then
  IMAGE_NAME=$2
fi

echo "building docker image: "
echo "  image: $IMAGE_NAME:$VERSION_TAG "
echo "  qube version: $VERSION_TAG"
echo

echo "docker build --no-cache --pull --build-arg=COMMIT=$(git rev-parse --short HEAD) --build-arg=QUBES_VERSION=$VERSION_TAG --build-arg=BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -t $IMAGE_NAME:$VERSION_TAG ."
docker build --no-cache --pull --build-arg=COMMIT=$(git rev-parse --short HEAD) --build-arg=QUBES_VERSION=$VERSION_TAG --build-arg=BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -t $IMAGE_NAME:$VERSION_TAG .
