#!/bin/bash

VERSION="0.2"

rm -rf bin

CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -a -o bin/macos/jongleur
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -a -o bin/linux/jongleur
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o bin/windows/jongleur.exe

tar czf bin/jongleur-$VERSION-macos.tar.gz --directory=bin/macos jongleur
tar czf bin/jongleur-$VERSION-linux.tar.gz --directory=bin/linux jongleur
zip     bin/jongleur-$VERSION-windows.zip -j bin/windows/jongleur.exe

mkdir bin/docker
cp Dockerfile bin/docker/Dockerfile
cp bin/linux/jongleur bin/docker/jongleur

docker build --no-cache -t "maxmanuylov/jongleur:$VERSION" bin/docker
docker tag "maxmanuylov/jongleur:$VERSION" "maxmanuylov/jongleur:latest"

#docker push "maxmanuylov/jongleur:$VERSION"
#docker push "maxmanuylov/jongleur:latest"
