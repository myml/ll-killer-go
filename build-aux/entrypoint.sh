#!/bin/bash

if [ -e ".killer-debug" ];then
    export KILLER_DEBUG=1
fi

APP_DIR=${APP_DIR:-"/opt/apps/$LINGLONG_APPID/files"}
OVERLAY_EXEC_PATH=$APP_DIR/fuse-overlayfs
if [ ! -e "$OVERLAY_EXEC_PATH" ];then
    OVERLAY_EXEC_PATH=$(which "fuse-overlayfs")
fi
OVERLAY_EXEC=${OVERLAY_EXEC:-"$OVERLAY_EXEC_PATH"}
KILLER_EXEC=${KILLER_EXEC:-"$APP_DIR/ll-killer"}
if (test -z "$NO_OVERLAYFS" && "$OVERLAY_EXEC" --version && test -e /run/host/rootfs/dev/fuse ) 2>/dev/null >/dev/null; then
    exec $APP_DIR/ll-killer exec \
        --fuse-overlayfs "$OVERLAY_EXEC" \
        --fuse-overlayfs-args "$OVERLAY_ARGS" \
        --mount "/run/host/rootfs/dev:/dev:rbind" \
        --mount "/:/run/app.oldfs:rbind" \
        --mount "overlay:$APP_DIR/usr/share::fuse-overlayfs:lowerdir=$APP_DIR/share:/run/app.oldfs/usr/share,squash_to_root,static_nlink" \
        --mount "overlay:/run/app.rootfs::fuse-overlayfs:lowerdir=$APP_DIR:/run/app.oldfs,squash_to_root" \
        --mount "/proc:/run/app.rootfs/proc:rbind" \
        --mount "/dev:/run/app.rootfs/dev:rbind" \
        --mount "/run:/run/app.rootfs/run:rbind" \
        --mount "/sys:/run/app.rootfs/sys:rbind" \
        --mount "/tmp:/run/app.rootfs/tmp:rbind" \
        --mount "/home:/run/app.rootfs/home:rbind" \
        --mount "/root:/run/app.rootfs/root:rbind" \
        --rootfs /run/app.rootfs \
        --socket=/run/app.unix \
        -- "${@:-bash}"
fi
exec $APP_DIR/ll-killer exec \
    --mount "$APP_DIR/share:$APP_DIR/usr/share:rbind" \
    --mount "/+$APP_DIR:/run/app.rootfs::merge" \
    --mount "/run/host/rootfs/dev:/run/app.rootfs/dev:rbind" \
    --mount "/proc:/run/app.rootfs/proc:rbind" \
    --mount "/run:/run/app.rootfs/run:rbind" \
    --mount "/sys:/run/app.rootfs/sys:rbind" \
    --mount "/tmp:/run/app.rootfs/tmp:rbind" \
    --mount "/home:/run/app.rootfs/home:rbind" \
    --mount "/root:/run/app.rootfs/root:rbind" \
    --socket=/run/app.unix \
    --rootfs=/run/app.rootfs \
    -- "${@:-bash}"