#!/bin/bash
source $(dirname $0)/env.sh
setup-filesystem.sh
mkdir -p "$PREFIX/bin"
cp -af "$KILLER_EXEC" "build-aux/$ENTRYPOINT" "$PREFIX"
ln -sf "$PREFIX/$ENTRYPOINT" "$PREFIX/bin/$ENTRYPOINT"
mv "$PREFIX/usr/share" "$PREFIX/share" || mkdir -p $PREFIX/share
mkdir -p "$PREFIX/usr/share"
cp -arfLT $PREFIX/opt/apps/*/entries $PREFIX/share 
cp -arfLT $PREFIX/opt/apps/*/files/share $PREFIX/share
find $PREFIX/share -xtype l -exec "relink.sh" "{}" \;
find $PREFIX/share/applications -name "*.desktop" -exec "setup-desktop.sh" "{}" \; 
