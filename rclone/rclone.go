package rclone

import (
	"clonedrive/mounter"
)

// Rclone struct
type Rclone struct {
	path  string
	Mount mounter.Mount
	Move  Move
}

// New constructor
func New(src string, dst string, cache string) *Rclone {
	// Call the OS-specific mount constructor
	mount := mounter.Mounter(src+":", dst)

	mount.Overlay = false
	mount.Mounter = []string{
		RclonePath,
		"mount",
		"--read-only",
		"--allow-other",
		"--no-modtime",
		"--dir-cache-time=240m",
		"--tpslimit=10",
		"--tpslimit-burst=1",
		"--buffer-size=1G",
		src + ":",
		dst,
	}

	return &Rclone{
		path:  RclonePath,
		Mount: *mount,
		Move:  *Mover(cache, src),
	}
}
