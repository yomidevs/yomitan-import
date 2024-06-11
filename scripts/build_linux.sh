#!/bin/bash

build_dir="yomitan-import-linux"

mkdir -p "$build_dir"

go build -o "yomitan-import-linux" ./yomitan
go build -o "yomitan-import-linux" ./yomitan-gtk

zip -r "$build_dir.zip" "$build_dir"
