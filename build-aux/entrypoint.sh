#!/bin/bash
APP_DIR="/opt/apps/$LINGLONG_APPID/files"
$APP_DIR/ll-killer exec \
    --mount "$APP_DIR/share:$APP_DIR/usr/share:rbind" \
    --mount "/+$APP_DIR:/run/app.rootfs::merge" \
    --socket=/run/app.unix \
    --rootfs=/run/app.rootfs \
    -- "${@:-bash}"
