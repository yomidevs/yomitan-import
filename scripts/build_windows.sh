#!/bin/bash

build_dir="yomitan-import-windows"
mkdir -p "$build_dir"

CGO_ENABLED=1 CC=/usr/bin/x86_64-w64-mingw32-gcc CXX=/usr/bin/x86_64-w64-mingw32-g++ GOOS=windows OARCH=amd64 go build -o "yomitan-import-windows" -ldflags '-extldflags "-static"' ./yomitan
CGO_ENABLED=1 CC=/usr/bin/x86_64-w64-mingw32-gcc CXX=/usr/bin/x86_64-w64-mingw32-g++ GOOS=windows OARCH=amd64 go build -o "yomitan-import-windows" -ldflags '-extldflags "-static"' ./yomitan-gtk

zip -r "$build_dir.zip" "$build_dir"
