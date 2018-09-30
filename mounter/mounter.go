package mounter

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// Mount struct
type Mount struct {
	Mounter    []string
	unmounter  []string
	useChecker []string
	source     string
	mountPoint string
	ready      chan bool
	Overlay    bool
	Kill       chan bool
	killed     bool
}

func (m *Mount) unmount() bool {
	command := exec.Command(string(m.unmounter[0]), m.unmounter[1:]...)
	if err := command.Run(); err != nil {
		log.Fatalf("unmount: Run: %v", err)
		return false
	}

	return true
}

func (m *Mount) inUse() bool {
	command := exec.Command(string(m.useChecker[0]), m.useChecker[1:]...)
	if err := command.Start(); err != nil {
		log.Fatalf("inUse: Start: %v", err)
	}
	if err := command.Wait(); err != nil {
		return false
	}

	return true
}

// Mount method
func (m *Mount) Mount() {
	go func(m *Mount) {
		fmt.Printf("Waiting for mount kill command\n")
		m.killed = <-m.Kill
		fmt.Printf("Mount kill command received\n")
		m.unmount()
	}(m)

	for {
		// Wait for a ready signal on overlay mounts
		if m.Overlay == true {
			// Any signal will do
			<-m.ready
		}

		// Make sure nothing's mounted
		if m.inUse() {
			m.unmount()
		}

		// This is a new mount session
		m.killed = false

		// If not in use
		if !m.inUse() {
			// Run the command, sleep briefly, signal the overlay and wait
			log.Print("Trying the rclone mount")
			command := exec.Command(string(m.Mounter[0]), m.Mounter[1:]...)

			log.Printf("Args: %v", command.Args)
			if err := command.Start(); err != nil {
				log.Fatalf("mount: Start: %v", err)
			}

			log.Print("Sleeping")
			time.Sleep(3 * time.Second)

			log.Print("Overlay ready?")
			if m.Overlay == false {
				m.ready <- true
			}

			log.Print("Waiting for command to complete")
			command.Wait()

			if m.killed {
				break
			}
		}
	}
}
