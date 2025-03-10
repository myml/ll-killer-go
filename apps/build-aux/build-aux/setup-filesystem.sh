#!/bin/bash

cp -arfT linglong/filesystem/diff "$PREFIX";
find $PREFIX \( -type c -or -name ".wh.*" \) -exec rm -rf {} \; ;
mv "$PREFIX/usr/share" "$PREFIX/share";
mkdir $PREFIX/share;
