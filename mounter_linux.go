package main

import (
	"fmt"
	"os/exec"
	"os/user"
	"path"
)

var rclonePath = rclonePath()

func rclonePath() string {
	// Find the home directory
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	return path.Join(homeDir, "bin", "rclone")
}

// Mounter constructor for Linux
func Mounter(src string, dst string) *Mounter {
	newMount := Mount{source: src, mountPoint: dst}

	newMount.unmounter = *exec.Command("/usr/bin/sudo", "/usr/bin/umount", newMount.mountPoint)
	newMount.useChecker = *exec.Command("/sbin/lsof", newMount.mountPoint)

	newMount.ready = make(chan bool, 1)

	return &newMount
}

func overlayMount(cacheDir string, localDir string, dst string) *Mounter {
	// Call the OS-specific mount constructor
	mounter := newMounter("", dst)

	mounter.overlay = true
	mounter.mounter = *exec.Command("/usr/bin/sudo", "/usr/bin/mount", mounter.mountPoint)

	return mounter
}
