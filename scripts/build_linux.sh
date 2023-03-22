#!/bin/bash

mkdir -p dst
mkdir -p yomitan-import

go build github.com/themoeway/yomitan-import/tree/master/yomichan
go build github.com/themoeway/yomitan-import/tree/master/yomichan-gtk

mv yomitan yomitan-import
mv yomitan-gtk yomitan-import

tar czvf dst/yomitan-import_linux.tar.gz yomitan-import

rm -rf yomitan-import
