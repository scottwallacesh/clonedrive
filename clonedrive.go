package main

import (
	"log"
	"os/exec"
	"os/user"
	"path"
	"time"
)

type mounter struct {
	mounter    exec.Cmd
	unmounter  exec.Cmd
	useChecker exec.Cmd
	source     string
	mountPoint string
	ready      chan bool
	overlay    bool
}

func (m *mounter) unmount() bool {
	if err := m.unmounter.Run(); err != nil {
		log.Fatalf("unmount: Run: %v", err)
		return false
	}

	return true
}

func (m *mounter) inUse() bool {
	if err := m.useChecker.Start(); err != nil {
		log.Fatalf("inUse: Start: %v", err)
	}
	if err := m.useChecker.Wait(); err != nil {
		return false
	}

	return true
}

func (m *mounter) mount() {
	for {
		// Wait for a ready signal on overlay mounts
		if m.overlay == true {
			// Any signal will do
			<-m.ready
		}

		// Make sure nothing's mounted
		m.unmount()

		// If not in use
		if !m.inUse() {
			// Run the command, sleep briefly, signal the overlay and wait
			if err := m.mounter.Start(); err != nil {
				log.Fatalf("mount: Start: %v", err)
			}

			time.Sleep(3 * time.Second)

			if m.overlay == false {
				m.ready <- true
			}

			m.mounter.Wait()
		}
	}
}

func rcloneMount(src string, dst string) *mounter {
	// Call the OS-specific mount constructor
	mounter := newMounter(src+":", dst)

	mounter.overlay = false
	mounter.mounter = *exec.Command(rclonePath, "mount",
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

	rclone := rcloneMount(remoteDrive, localDir)
	overlay := overlayMount(cacheDir, localDir, overlayDir)
}
