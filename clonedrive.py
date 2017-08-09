#!/usr/bin/env python

import subprocess
import os
import threading
import ctypes.util
import time

def convert_sleeptime(timestring):
    """Function to convert a simply timespan string to number of seconds."""
    seconds_per_unit = {"s": 1, "m": 60, "h": 3600, "d": 86400, "w": 604800}

    try:
        return int(timestring)
    except ValueError:
        try:
            return int(timestring[:-1]) * seconds_per_unit[timestring[-1]]
        except KeyError:
            raise
            return None


def unmount(directory):
    """Function to unmount a directory."""
    libc = ctypes.util.find_library('c')
    libc.umount(directory)


def directory_in_use(directory):
    """Simple function to check if a directory
       is in use."""
    lsof = subprocess.Popen(['/usr/sbin/lsof', directory])
    lsof.wait()

    if lsof.returncode == 1:
        return False
    else:
        return True


def rclone_mounter(rclone_remote, directory):
    """Simple function to mount rclone remote."""
    while True:
        unmount(directory)
        if not directory_in_use(directory):
            rclone = subprocess.Popen(['/usr/local/bin/rclone',
                                       'mount',
                                       '--read-only',
                                       '--allow-others',
                                       '--no-modtime',
                                       '--dir-cache-time=10s',
                                       '--buffer-size=1G',
                                       '%s:' % rclone_remote,
                                       directory
                                       ])
            rclone.wait()


def unionfs_mounter(sourcelist=[], directory=None):
    """Simple function to mount a unionfs 'stack'."""
    source = ':'.join([directory + '=' + readwrite
                          for (directory, readwrite) in sourcelist])

    while True:
        unmount(directory)
        if not directory_in_use(directory):        
            union = subprocess.Popen(['/usr/local/bin/unionfs',
                                      '-f',
                                      '-o', 'cow',
                                      source,
                                      directory
                                      ])
            union.wait()


def rclone_mover(directory, rclone_remote, sleeptime='6h', schedule=None):
    """Simple function to move directory contents to rclone remote."""
    while True:
        os.chdir(directory)

        command = ['/usr/local/bin/rclone',
                   'move',
                   '.',
                   '%s:' % rclone_remote
                   ]

        if schedule:
            command.append('--bwlimit=%s' % schedule)

        rclone = subprocess.Popen(command)

        rclone.wait()

        time.sleep(convert_sleeptime(sleeptime))


if __name__ == '__main__':
    remote_drive = 'GoogleDriveCrypt'
    local_mount = os.path.expanduser('~/mnt/GoogleDriveCrypt')
    cache_drive = os.path.expanduser('~/mnt/cache')
    master_mount = os.path.expanduser('/Volumes/Download')
    union_mount = os.path.expanduser('~/mnt/union')

    rclone_mount = threading.Thread(target=rclone_mounter,
                                    args=(remote_drive, local_mount)
                                    )
    
    unionfs_mount = threading.Thread(target=unionfs_mounter,
                                     args=([(local_mount, 'RO'),
                                            (master_mount, 'RO'),
                                            (cache_drive, 'RW')
                                            ],
                                            union_mount)
                                     )

    rclone_move = threading.Thread(target=rclone_mover,
                                   args=(cache_drive,
                                         remote_drive,
                                         schedule='07:00,1M 23:00,off')
                                   )

    rclone_mount.start()
    unionfs_mount.start()
    rclone_move.start()

    try:
        while True:
            for thread in [rclone_mount, unionfs_mount, rclone_move]:
                if thread.is_alive():
                    thread.join(0.5)
    except KeyboardInterrupt:
        sys.exit(0)