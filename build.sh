#!/bin/bash

VERSION="v0.9"

rm -rf bin

export GO15VENDOREXPERIMENT=1

GOOS=darwin  GOARCH=amd64 go build -o bin/macos/jongleur
GOOS=linux   GOARCH=amd64 go build -o bin/linux/jongleur
GOOS=windows GOARCH=amd64 go build -o bin/windows/jongleur.exe

tar czf bin/jongleur-$VERSION-macos.tar.gz --directory=bin/macos jongleur
tar czf bin/jongleur-$VERSION-linux.tar.gz --directory=bin/linux jongleur
zip     bin/jongleur-$VERSION-windows.zip -j bin/windows/jongleur.exe
