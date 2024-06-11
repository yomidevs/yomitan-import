#!/bin/bash

build_dir="yomitan-import-windows"

mkdir -p "$build_dir"

CGO_ENABLED=1 CC=/usr/bin/i686-w64-mingw32-g++ GOOS=windows OARCH=x64 go build -o "yomitan-import-windows" ./yomitan
CGO_ENABLED=1 CC=/usr/bin/i686-w64-mingw32-g++ GOOS=windows OARCH=x64 go build -o "yomitan-import-windows" ./yomitan-gtk

zip -r "$build_dir.zip" "$build_dir"
