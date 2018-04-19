package mounter

import (
	"fmt"
	"os/exec"
)

// Mounter constructor for Darwin
func Mounter(src string, dst string) *Mount {
	newMount := Mount{source: src, mountPoint: dst}

	newMount.unmounter = *exec.Command("/usr/sbin/diskutil", "unmount", newMount.mountPoint)
	newMount.useChecker = *exec.Command("/usr/sbin/lsof", newMount.mountPoint)

	newMount.ready = make(chan bool, 1)

	return &newMount
}

// OVerlayMounter constructor for Darwin
func OverlayMount(cacheDir string, localDir string, dst string) *Mount {
	// Call the OS-specific mount constructor
	src := fmt.Sprintf("%s=%s:%s=%s", cacheDir, "RW",
		localDir, "RO")
	mounter := Mounter(src, dst)

	mounter.Overlay = true
	mounter.Mounter = *exec.Command("/usr/local/bin/unionfs",
		"-o", "cow,direct_io,auto_cache", mounter.source, mounter.mountPoint)

	return mounter
}
