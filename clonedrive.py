#!/usr/bin/env python

import subprocess
import os
from multiprocessing import Process, Pipe
import ctypes
import ctypes.util
import time
import sys


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
    umounter = subprocess.Popen(['/usr/bin/sudo',
                                 '/usr/bin/umount',
                                 directory
                                 ])
    umounter.wait()


def directory_in_use(directory):
    """Function to check if a directory is in use."""
    lsof = subprocess.Popen(['/sbin/lsof', directory],
                            stdout=None,
                            stderr=None)
    lsof.wait()

    if lsof.returncode == 1:
        return False
    else:
        return True


def rclone_mounter(rclone_remote, directory, pipe):
    """Function to mount rclone remote."""
    homedir = os.path.expanduser('~')
    rclone_bin = os.path.join(homedir, 'bin', 'rclone')
    while True:
        # Umount the existing directory, just in case
        unmount(directory)
        if not directory_in_use(directory):
            # Mount it
            rclone = subprocess.Popen([rclone_bin,
                                       'mount',
                                       '--read-only',
                                       '--allow-other',
                                       '--no-modtime',
                                       '--dir-cache-time=240m',
                                       '--tpslimit=10',
                                       '--tpslimit-burst=1',
                                       '--buffer-size=1G',
                                       '%s:' % rclone_remote,
                                       directory
                                       ])

            # Wait a few seconds for the mount to complete
            time.sleep(3)

            # Send a signal to the overlay_mounter thread
            pipe.send(True)

            # Wait for rclone to exit
            rclone.wait()


def overlay_mounter(directory, pipe):
    """Function to mount a overlay 'stack'."""
    while True:
        # Wait for a signal from the rclone_mounter thread
        if pipe.recv() == True:
            # Umount the existing directory, just in case
            unmount(directory)
            if not directory_in_use(directory):
                # Mount it
                union = subprocess.Popen(['/usr/bin/sudo',
                                          '/usr/bin/mount',
                                          directory
                                          ])


def rclone_mover(directory, rclone_remote, sleeptime='6h', schedule=None):
    """Function to move cache directory contents to rclone remote."""
    homedir = os.path.expanduser('~')
    rclone_bin = os.path.join(homedir, 'bin', 'rclone')
    while True:
        # Build the command line
        command = [rclone_bin,
                   'move',
                   '.',
                   '%s:' % rclone_remote,
                   '--exclude=.unionfs'
                   ]

        # Append the schedule, if appropriate
        if schedule:
            command.append('--bwlimit=%s' % schedule)

        # Run the command
        rclone = subprocess.Popen(command, cwd=directory)
        rclone.wait()

        # Sleep until the next schedule
        time.sleep(convert_sleeptime(sleeptime))


if __name__ == '__main__':
    # Main directories
    remote_drive = 'GoogleDriveCrypt'
    homedir = os.path.expanduser('~')
    local_dir = os.path.join(homedir, 'mnt', 'GoogleDriveCrypt')
    overlay_dir = os.path.join(homedir, 'mnt', 'union')
    cache_dir = os.path.join(homedir, 'mnt', 'cache')

    # Create a cross-thread pipe
    rclone_pipe, overlay_pipe = Pipe()

    # Prepare the threads
    rclone_mount = Process(target=rclone_mounter,
                           args=(remote_drive, local_dir, rclone_pipe)
                           )

    overlay_mount = Process(target=overlay_mounter,
                            args=(overlay_dir, overlay_pipe)
                            )

    rclone_move = Process(target=rclone_mover,
                          args=(cache_dir,
                                remote_drive,
                                '6h',
                                '07:00,1M 23:00,off')
                          )

    # Start the threads
    rclone_mount.start()
    overlay_mount.start()
    rclone_move.start()

    # Wait for a keyboard interrupt
    try:
        while True:
            for thread in [rclone_mount, overlay_mount, rclone_move]:
                if thread.is_alive():
                    thread.join(0.5)
    except KeyboardInterrupt:
        # Kill the threads
        overlay_mount.terminate()
        rclone_mount.terminate()

        # Umount the filesystems
        unmount(overlay_dir)
        unmount(local_dir)

        # Clean exit
        sys.exit(0)
