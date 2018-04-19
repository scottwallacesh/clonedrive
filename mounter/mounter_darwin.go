package mounter

import (
	"fmt"
)

// Mounter constructor for Darwin
func Mounter(src string, dst string) *Mount {
	newMount := Mount{source: src, mountPoint: dst}

	newMount.unmounter = []string{"/usr/sbin/diskutil", "unmount", newMount.mountPoint}
	newMount.useChecker = []string{"/usr/sbin/lsof", newMount.mountPoint}

	newMount.ready = make(chan bool, 1)

	return &newMount
}

// OverlayMount constructor for Darwin
func OverlayMount(cacheDir string, localDir string, dst string) *Mount {
	// Call the OS-specific mount constructor
	src := fmt.Sprintf("%s=%s:%s=%s", cacheDir, "RW",
		localDir, "RO")
	mounter := Mounter(src, dst)

	mounter.Overlay = true
	mounter.Mounter = []string{
		"/usr/local/bin/unionfs",
		"-o",
		"cow,direct_io,auto_cache",
		mounter.source,
		mounter.mountPoint,
	}

	return mounter
}
