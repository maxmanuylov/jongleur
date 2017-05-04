#!/bin/bash

if [ -z "$VERSION" ]; then
    echo "VERSION is not specified" 1>&2
    exit 255
fi

OLD_WD="$PWD"
cd "$(dirname "$0")/.."

govendor sync

rm -rf bin

TEMP_GOPATH="$PWD/.go"
TEMP_PROJECT_PARENT_DIR="$TEMP_GOPATH/src/github.com/maxmanuylov"
TEMP_PROJECT_DIR="$TEMP_PROJECT_PARENT_DIR/jongleur"
PKG_DIR="$PWD/.pkg"

mkdir -p "$TEMP_PROJECT_PARENT_DIR"
ln -s "$PWD" "$TEMP_PROJECT_DIR"

cd "$TEMP_PROJECT_DIR"

export GOPATH="$TEMP_GOPATH"
export GOARCH=amd64

GOOS=darwin  go build -i -pkgdir "$PKG_DIR/darwin"  -o bin/macos/jongleur
GOOS=linux   go build -i -pkgdir "$PKG_DIR/linux"   -o bin/linux/jongleur
GOOS=windows go build -i -pkgdir "$PKG_DIR/windows" -o bin/windows/jongleur.exe

tar czf bin/jongleur-$VERSION-macos.tar.gz --directory=bin/macos jongleur
tar czf bin/jongleur-$VERSION-linux.tar.gz --directory=bin/linux jongleur
zip     bin/jongleur-$VERSION-windows.zip -j bin/windows/jongleur.exe

cat <<EOF >bin/env.txt
VERSION=$VERSION
BUILD=$BUILD
REVISION=$REVISION
EOF

echo "##message{\"kind\":\"Status\",\"text\":\"OK\"}"

rm -rf "$TEMP_GOPATH"

cd "$OLD_WD"
