#!/bin/bash

OLD_WD="$PWD"
cd "$(dirname "$0")"

GO_VERSION="1.8"
GO_BUILD_IMAGE="maxmanuylov/go-build"

sed -e "s|@GO_VERSION@|$GO_VERSION|g" Dockerfile.dist > Dockerfile

docker build --no-cache -t "$GO_BUILD_IMAGE:$GO_VERSION" .
docker push "$GO_BUILD_IMAGE:$GO_VERSION"

cd "$OLD_WD"
