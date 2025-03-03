#!/bin/bash
source $(dirname $0)/env.sh

echo "[准备文件系统]"
setup-filesystem.sh

echo "[复制必要文件]"
test -e "build-aux/fuse-overlayfs"&&cp -avf "build-aux/fuse-overlayfs" "$PREFIX"
cp -avf "$KILLER_EXEC" "$PREFIX/ll-killer"
cp -avf "build-aux/$ENTRYPOINT_NAME" "$PREFIX"

echo "[调整文件布局]"
mv "$PREFIX/usr/share" "$PREFIX/share" || mkdir -p $PREFIX/share
mkdir -p "$PREFIX/usr/share"

echo "[合并share目录]"
find $PREFIX/opt/apps/ -type d \( -path "$PREFIX/opt/apps/*/entries" \
        -or -path "$PREFIX/opt/apps/*/files/share" \) \
        -exec "move-share.sh" "{}" \;
# cp -avrfLT $PREFIX/opt/apps/*/entries $PREFIX/share 
# cp -avrfLT $PREFIX/opt/apps/*/files/share $PREFIX/share


echo "[修正符号链接]"
# https://github.com/OpenAtom-Linyaps/linyaps/issues/1039
find $PREFIX/share -xtype l -exec "relink.sh" "{}" \;

echo "[配置快捷方式]"
find $PREFIX/share/applications -name "*.desktop" -exec "setup-desktop.sh" "{}" \; 
