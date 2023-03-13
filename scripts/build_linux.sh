#!/bin/sh

mkdir -p dst
mkdir -p yomichan-import

# TODO: properly output as yomitan
go build foosoft.net/projects/yomichan-import/yomichan
go build foosoft.net/projects/yomichan-import/yomichan-gtk

mv yomichan yomichan-import
mv yomichan-gtk yomichan-import

tar czvf dst/yomichan-import_linux.tar.gz yomichan-import/yomichan yomichan-import/yomichan-gtk
