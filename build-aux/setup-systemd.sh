#!/bin/bash
set -e
source $(dirname $0)/env.sh
mkdir -p $PREFIX/share/systemd/user
SRC="$1"
DST="$PREFIX/share/systemd/user/$(basename $SRC)"
mv -Tv "$SRC" "$DST"
sed -i -E -e "s:^\s*ExecStart\s*=:ExecStart=$ENTRYPOINT :g" -e '/^User=/d' -e '/WantedBy/ s/multi-user.target/default.target/' "$DST"
RVL=$(realpath --relative-to="$(dirname "$SRC")" "$DST")
ln -svTf "$RVL" "$SRC"