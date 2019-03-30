#!/usr/bin/env python

"""Script to mount a read-only, encrypted Google Drive with a read-write,
   local cache for use with something like Plex.
"""

import logging
import multiprocessing
import os
import platform
import signal
import subprocess
import sys
import time


MAX_OVERLAY_RETRIES = 5


class Mounter(object):
    """Class for mounting filesystems."""

    def __init__(self, src, dst, pipe=None):
        """Initialisation method for the class.
           Determine the OS and configure a few (hardcoded!) paths.
        """
        if platform.system() == 'Linux':
            self.mount_bin = ['/usr/bin/sudo', '/usr/bin/mount']
            self.umount_bin = ['/usr/bin/sudo', '/usr/bin/umount']
            self.lsof_bin = '/sbin/lsof'
            self.rclone_bin = os.path.expanduser('~/bin/rclone')

        if platform.system() == 'Darwin':
            self.mount_bin = ['/sbin/mount']
            self.umount_bin = ['/usr/sbin/diskutil', 'unmount']
            self.lsof_bin = '/usr/sbin/lsof'
            self.rclone_bin = '/usr/local/bin/rclone'

        self.source = src
        self.mount_point = dst
        self.command = None

        if pipe is not None:
            # Use the given pipe, we must be an overlay FS
            self.parent_pipe = None
            self.child_pipe = pipe
        else:
            # Create a pipe for signalling to the overlay
            # that the mount is ready to be overlaid.
            self.parent_pipe, self.child_pipe = multiprocessing.Pipe()

    def set_command(self, command):
        """ Setter method. """
        self.command = command

    def mount(self):
        """Method to mount the filesystem."""
        while True:
            # Only wait for signal if we're the overlay
            if self.parent_pipe is None:
                # Any signal will do
                self.child_pipe.recv()

            # Unmount, to be safe
            self.unmount()

            # Check the mountpoint isn't being used
            if not self.in_use():
                try:
                    # Ensure the command is legit
                    command = filter(None,
                                     self.command +
                                     [self.source, self.mount_point])

                    # Mount the mountpoint
                    mount = subprocess.Popen(command,
                                             stdout=subprocess.PIPE,
                                             stderr=subprocess.PIPE)

                    # Wait a few seconds
                    time.sleep(3)

                    # Send a signal if we're not the overlay
                    if self.parent_pipe is not None:
                        self.parent_pipe.send(True)

                    # Wait until the mount stops
                    (out, err) = mount.communicate()
                except OSError as errmsg:
                    logging.error('%s: %s', self.command[0], errmsg)
                    logging.error(out)
                    logging.error(err)
                    break

    def unmount(self):
        """Method to unmount."""
        unmounter = subprocess.Popen(self.umount_bin + [self.mount_point],
                                     stdout=subprocess.PIPE,
                                     stderr=subprocess.PIPE)
        (out, err) = unmounter.communicate()

        if err:
            logging.error(err)
        if out:
            logging.info(out)

    def in_use(self):
        """Method to check if a directory is in use."""
        try:
            lsof = subprocess.Popen([self.lsof_bin, self.mount_point],
                                    stdout=subprocess.PIPE,
                                    stderr=subprocess.PIPE)
        except OSError as errmsg:
            logging.error('%s: %s', errmsg, lsof.stderr.read())
            return None

        (out, err) = lsof.communicate()

        if lsof.returncode == 1:
            if out:
                logging.warning(out)
            if err:
                logging.warning(err)
            return False

        return True

    def has_files(self):
        """Method to test if files are detected in the mountpoint."""
        if os.listdir(self.mount_point):
            return True
        return False


class RcloneMounter(Mounter):
    """Class for mounting rclone filesystems."""

    def __init__(self, *args, **kwargs):
        super(RcloneMounter, self).__init__(*args, **kwargs)
        self.set_command([self.rclone_bin,
                          'mount',
                          '--read-only',
                          '--allow-other',
                          '--no-modtime',
                          '--dir-cache-time=240m',
                          '--tpslimit=10',
                          '--tpslimit-burst=1',
                          '--buffer-size=1G'])

        # Append a colon to the source
        self.source += ':'


class UnionfsMounter(Mounter):
    """Class for mounting union filesystems."""

    def __init__(self, *args, **kwargs):
        super(UnionfsMounter, self).__init__(*args, **kwargs)
        self.set_command(['/usr/local/bin/unionfs',
                          '-o', 'cow,direct_io,auto_cache'])


class OverlayMounter(Mounter):
    """Class for mounting overlayfs filesystems."""

    def __init__(self, *args, **kwargs):
        super(OverlayMounter, self).__init__(*args, **kwargs)
        self.set_command(self.mount_bin)


def kill_and_unmount(threads, overlay, rclone):
    """Function to wind up the threads and unmount the mountpoints."""
    # Kill the threads
    for thread in threads:
        logging.info('Terminating %s', thread.name)
        thread.terminate()

    # Wait for the threads
    for thread in threads:
        logging.info('Waiting for %s to finish', thread.name)
        thread.join()

    # Umount the filesystems
    overlay.unmount()
    rclone.unmount()

def main():
    """Function to call the main programm logic."""

    # Main directories
    remote_drive = 'GoogleDriveCrypt'
    homedir = os.path.expanduser('~')
    local_dir = os.path.join(homedir, 'mnt', 'GoogleDriveCrypt')
    overlay_dir = os.path.join(homedir, 'mnt', 'union')
    cache_dir = os.path.join(homedir, 'mnt', '.cache')

    # Rclone mounter
    rclone = RcloneMounter(remote_drive, local_dir)

    # Overlay mounter
    if platform.system() == 'Linux':
        overlay = OverlayMounter(None, overlay_dir, rclone.child_pipe)

    elif platform.system() == 'Darwin':
        source = '%s=%s:%s=%s' % (cache_dir, 'RW',
                                  local_dir, 'RO')

        overlay = UnionfsMounter(source, overlay_dir, rclone.child_pipe)

    # Wait for a keyboard interrupt
    try:
        # Main thread loop
        while True:
            logging.info('Attempting to mount rclone and overlay filesystems')

            # Prepare the threads
            threads = []
            threads.append(multiprocessing.Process(target=rclone.mount,
                                                   name='rclone mount'))
            threads.append(multiprocessing.Process(target=overlay.mount,
                                                   name='overlay mount'))

            while True:
                for thread in threads:
                    if not thread.is_alive():
                        logging.info('Starting %s', thread.name)
                        thread.start()
                    thread.join(2.0)

                # Register a SIGTERM handler here
                signal.signal(signal.SIGTERM,
                              kill_and_unmount(threads, overlay, rclone))

                # Check that the overlay mount contains files
                retries = 0
                while retries < MAX_OVERLAY_RETRIES:
                    if overlay.has_files():
                        break

                    logging.info('Overlay filesystem contains no files.')
                    retries += 1

                    logging.info('Retry: %d of %d.  Waiting for %d seconds...',
                                 retries, MAX_OVERLAY_RETRIES, (retries * 5))
                    time.sleep(retries * 5)
                else:
                    # Try unmounting the filesystem to start over
                    logging.info('Retries exhausted.  Attempting to remount.')
                    kill_and_unmount(threads, overlay, rclone)
                    break

                if retries > 0:
                    logging.info('Files detected in the overlay filesystem.')
                    logging.info('Resuming normal operation.')

    except KeyboardInterrupt:
        kill_and_unmount(threads, overlay, rclone)


if __name__ == '__main__':
    logging.basicConfig(format='%(asctime)-15s %(levelname)s: %(message)s',
                        level=logging.INFO)
    main()

    # Clean exit
    sys.exit(0)
