package main

import (
	"fmt"
	"os/exec"
)

var rclonePath = "/usr/local/bin/rclone"

func newMounter(src string, dst string) *mounter {
	newMount := mounter{source: src, mountPoint: dst}

	newMount.unmounter = *exec.Command("/usr/sbin/diskutil", "unmount", newMount.mountPoint)
	newMount.useChecker = *exec.Command("/usr/sbin/lsof", newMount.mountPoint)

	newMount.ready = make(chan bool, 1)

	return &newMount
}

func overlayMount(cacheDir string, localDir string, dst string) *mounter {
	// Call the OS-specific mount constructor
	src := fmt.Sprintf("%s=%s:%s=%s", cacheDir, "RW",
		localDir, "RO")
	mounter := newMounter(src, dst)

	mounter.overlay = true
	mounter.mounter = *exec.Command("/usr/local/bin/unionfs",
		"-o", "cow,direct_io,auto_cache", mounter.source, mounter.mountPoint)

	return mounter
}
