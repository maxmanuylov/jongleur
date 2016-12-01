#!/bin/bash

VERSION="v0.8"
DOCKER_REGISTRY="docker.io"

rm -rf bin

export GO15VENDOREXPERIMENT=1

GOOS=darwin  GOARCH=amd64 go build -o bin/macos/jongleur
GOOS=linux   GOARCH=amd64 go build -o bin/linux/jongleur
GOOS=windows GOARCH=amd64 go build -o bin/windows/jongleur.exe

#tar czf bin/jongleur-$VERSION-macos.tar.gz --directory=bin/macos jongleur
#tar czf bin/jongleur-$VERSION-linux.tar.gz --directory=bin/linux jongleur
#zip     bin/jongleur-$VERSION-windows.zip -j bin/windows/jongleur.exe

mkdir bin/docker
cp Dockerfile bin/docker/Dockerfile
cp bin/linux/jongleur bin/docker/jongleur

docker build --no-cache -t "$DOCKER_REGISTRY/maxmanuylov/jongleur:$VERSION" bin/docker
docker tag "$DOCKER_REGISTRY/maxmanuylov/jongleur:$VERSION" "$DOCKER_REGISTRY/maxmanuylov/jongleur:latest"

#docker push "$DOCKER_REGISTRY/maxmanuylov/jongleur:$VERSION"
#docker push "$DOCKER_REGISTRY/maxmanuylov/jongleur:latest"
