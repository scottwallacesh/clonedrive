#!/bin/bash

MASTER_MOUNT=/Volumes/Downloads
GOOGL_MOUNT=~/mnt/GoogleDriveCrypt
CACHE_MOUNT=~/mnt/cache
UNION_MOUNT=~/mnt/union

# Simple function
function in_use() {
    MOUNT=${!}

    lsof ${MOUNT} > /dev/null 2>&1

    return ${?}
}

# Ensure the mountpoints exist
mkdir -p ${CACHE_MOUNT} ${GOOGL_MOUNT} ${UNION_MOUNT}

# Try to umount the mountpoints to reset and exit if not
umount ${UNION_MOUNT} ${GOOGL_MOUNT} 2> /dev/null
in_use ${UNION_MOUNT} || exit 1
in_use ${GOOGL_MOUNT} || exit 2


# Mount the directories
rclone mount --read-only \
             --dir-cache-time=10s \
             --buffer-size=1G \
             GoogleDriveCrypt: ${GOOGL_MOUNT} &
unionfs -o cow ${GOOGL_MOUNT}=RO:${MASTER_MOUNT}=RO:${CACHE_MOUNT}=RW ${UNION_MOUNT}

# Ensure local cache to moved to The Cloud regularly
while true; do
    cd ${CACHE_MOUNT} && rclone move . GoogleDriveCrypt:
    sleep 21600 # 4 hours
done
