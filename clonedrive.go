package main

import (
	"./mounter"
	"./rclone"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"
	"time"
)

func main() {
	remoteDrive := "GoogleDriveCrypt"

	// Find the home directory
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	// Create the path variables
	localDir := path.Join(homeDir, "mnt", remoteDrive)
	cacheDir := path.Join(homeDir, "mnt", "cache")
	overlayDir := path.Join(homeDir, "mnt", "union")

	// Prepare the mounts
	rclone := rclone.New(remoteDrive, localDir, cacheDir)
	overlay := mounter.OverlayMount(cacheDir, localDir, overlayDir)

	// Set the schedule for the mover
	rclone.Move.SetSchedule("07:00,1M 23:00,off")

	// Channel to handle OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGQUIT)

	// Goroutine to work with signals
	go func() {
		for sig := range c {
			if sig == syscall.SIGINT {
				overlay.Kill <- true
				rclone.Mount.Kill <- true
			}
			if sig == syscall.SIGQUIT {
				overlay.Kill <- true
				rclone.Mount.Kill <- true
				rclone.Move.Kill <- true
			}
		}
	}()

	// Main program loop
	for {
		// Launch the mounts and mover
		go rclone.Mount.Mount()
		go overlay.Mount()

		for {
			rclone.Move.Run()
			if rclone.Move.Killed {
				break
			}
			time.Sleep(rclone.Move.SleepTime)
		}

		if rclone.Move.Killed {
			break
		}
	}
}
