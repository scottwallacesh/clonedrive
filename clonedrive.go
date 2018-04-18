package main

import (
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"syscall"
	"time"
)

func rcloneMount(src string, dst string) *Mount {
	// Call the OS-specific mount constructor
	mounter := Mounter(src+":", dst)

	mounter.overlay = false
	mounter.mounter = *exec.Command(rclonePath,
		"mount",
		"--read-only",
		"--allow-other",
		"--no-modtime",
		"--dir-cache-time=240m",
		"--tpslimit=10",
		"--tpslimit-burst=1",
		"--buffer-size=1G")

	return mounter
}

func main() {
	remoteDrive := "GoogleDriveCrypt"

	// Find the home directory
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	// Create the path variables
	localDir := path.Join(homeDir, "mnt", remoteDrive)
	cacheDir := path.Join(homeDir, "mnt", "cache")
	overlayDir := path.Join(homeDir, "mnt", "union")

	// Prepare the mounts and mover
	rclone := rcloneMount(remoteDrive, localDir)
	overlay := overlayMount(cacheDir, localDir, overlayDir)
	rcloneMove := RcloneMover(cacheDir, remoteDrive)

	// Set the schedule for the mover
	rcloneMove.setSchedule("07:00,1M 23:00,off")

	// Channel to handle OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGQUIT)

	// Goroutine to work with signals
	go func() {
		for sig := range c {
			if sig == syscall.SIGINT {
				overlay.kill <- true
				rclone.kill <- true
			}
			if sig == syscall.SIGQUIT {
				overlay.kill <- true
				rclone.kill <- true
				rcloneMove.kill <- true
				os.Exit(0)
			}
		}
	}()

	// Main program loop
	for {
		// Launch the mounts and mover
		go rclone.mount()
		go overlay.mount()

		for {
			rcloneMove.mover.Run()
			if rcloneMove.killed {
				break
			}
			time.Sleep(rcloneMove.sleepTime)
		}
	}
}
