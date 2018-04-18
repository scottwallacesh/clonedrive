package rclone

import (
	"os/user"
	"path"
)

// RclonepPath for Linux
var RclonePath = rclonePath()

func rclonePath() string {
	// Find the home directory
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	return path.Join(homeDir, "bin", "rclone")
}
