package mounter

import (
	"os/exec"
)

// Mounter constructor for Linux
func Mounter(src string, dst string) *Mounter {
	newMount := Mount{source: src, mountPoint: dst}

	newMount.unmounter = *exec.Command("/usr/bin/sudo", "/usr/bin/umount", newMount.mountPoint)
	newMount.useChecker = *exec.Command("/sbin/lsof", newMount.mountPoint)

	newMount.ready = make(chan bool, 1)

	return &newMount
}

func OverlayMount(cacheDir string, localDir string, dst string) *Mounter {
	// Call the OS-specific mount constructor
	mounter := newMounter("", dst)

	mounter.Overlay = true
	mounter.Mounter = *exec.Command("/usr/bin/sudo", "/usr/bin/mount", mounter.mountPoint)

	return mounter
}
