package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"
	"time"

	"clonedrive/mounter"
	"clonedrive/rclone"
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
		fmt.Printf("Signal looper\n")
		for {
			fmt.Printf("Waiting for signal\n")
			for sig := range c {
				fmt.Printf("Signal received\n")
				if sig == syscall.SIGINT {
					fmt.Print("SIGINT received\n")
					overlay.Kill <- true
					rclone.Mount.Kill <- true
				}
				if sig == syscall.SIGQUIT {
					fmt.Print("SIGQUIT received\n")
					overlay.Kill <- true
					rclone.Mount.Kill <- true
					rclone.Move.Kill <- true
				}
				fmt.Printf("Signal completed\n")
				fmt.Printf("Awaiting further signals\n")
			}
		}
	}()

	// Main program loop
	for {
		// Launch the mounts and mover
		go rclone.Mount.Mount()
		go overlay.Mount()

		for {
			fmt.Printf("Running move\n")
			rclone.Move.Run()
			if rclone.Move.Killed {
				fmt.Printf("Move killed\n")
				break
			}
			time.Sleep(rclone.Move.SleepTime)
		}

		if rclone.Move.Killed {
			fmt.Printf("Move really killed\n")
			break
		}
	}
}
