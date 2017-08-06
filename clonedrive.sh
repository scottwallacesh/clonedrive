#!/bin/bash

MASTR_MOUNT=/Volumes/Downloads
GOOGL_MOUNT=~/mnt/GoogleDriveCrypt
CACHE_MOUNT=~/mnt/cache
UNION_MOUNT=~/mnt/union

# Simple function to test if a directory is in use
function in_use() {
    MOUNT=${1}

    lsof ${MOUNT} > /dev/null 2>&1

    return ${?}
}

# Ensure the mountpoints exist
mkdir -p ${CACHE_MOUNT} ${GOOGL_MOUNT} ${UNION_MOUNT}

while true; do
    # Try to umount the mountpoints to reset and exit if not
    umount ${UNION_MOUNT} ${GOOGL_MOUNT} 2> /dev/null
    in_use ${UNION_MOUNT} && exit 1
    in_use ${GOOGL_MOUNT} && exit 2

    # Mount Google Drive, read-only
    rclone mount --read-only \
                 --allow-other \
                 --no-modtime \
                 --dir-cache-time=10s \
                 --buffer-size=1G \
                 -v \
                 GoogleDriveCrypt: ${GOOGL_MOUNT} &

    # UnionFS magic
    unionfs -o cow ${GOOGL_MOUNT}=RO:${MASTR_MOUNT}=RO:${CACHE_MOUNT}=RW ${UNION_MOUNT}

    wait
done &

# Ensure local cache is moved to The Cloud regularly
while true; do
    cd ${CACHE_MOUNT} && rclone move . GoogleDriveCrypt: --bwlimit="07:00,1M 23:00,off"
    sleep 21600 # 6 hours
done

exit 0
