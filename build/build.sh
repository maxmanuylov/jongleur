#!/bin/bash

if [ -z "$VERSION" ]; then
    echo "VERSION is not specified" 1>&2
    exit 255
fi

OLD_WD="$PWD"
cd "$(dirname "$0")/.."

govendor sync

rm -rf bin

export GOARCH=amd64

GOOS=darwin  go build -o bin/macos/jongleur
GOOS=linux   go build -o bin/linux/jongleur
GOOS=windows go build -o bin/windows/jongleur.exe

tar czf bin/jongleur-$VERSION-macos.tar.gz --directory=bin/macos jongleur
tar czf bin/jongleur-$VERSION-linux.tar.gz --directory=bin/linux jongleur
zip     bin/jongleur-$VERSION-windows.zip -j bin/windows/jongleur.exe

cat <<EOF >bin/env.txt
BUILD=$BUILD
REVISION=$REVISION
EOF

echo "##message{\"kind\":\"Status\",\"text\":\"OK\"}"

cd "$OLD_WD"
