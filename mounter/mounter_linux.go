package mounter

// Mounter constructor for Linux
func Mounter(src string, dst string) *Mount {
	newMount := Mount{source: src, mountPoint: dst}

	newMount.unmounter = []string{"/usr/bin/sudo", "/usr/bin/umount", newMount.mountPoint}
	newMount.useChecker = []string{"/usr/bin/lsof", newMount.mountPoint}

	newMount.ready = make(chan bool, 1)

	return &newMount
}

// OverlayMount constructor for Linux
func OverlayMount(cacheDir string, localDir string, dst string) *Mount {
	// Call the OS-specific mount constructor
	mounter := Mounter("", dst)

	mounter.Overlay = true
	mounter.Mounter = []string{"/usr/bin/sudo", "/usr/bin/mount", mounter.mountPoint}

	return mounter
}
