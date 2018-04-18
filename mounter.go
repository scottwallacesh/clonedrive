package main

import (
	"log"
	"os/exec"
	"time"
)

// Mount struct
type Mount struct {
	mounter    exec.Cmd
	unmounter  exec.Cmd
	useChecker exec.Cmd
	source     string
	mountPoint string
	ready      chan bool
	overlay    bool
	kill       chan bool
	killed     bool
}

func (m *Mount) unmount() bool {
	if err := m.unmounter.Run(); err != nil {
		log.Fatalf("unmount: Run: %v", err)
		return false
	}

	return true
}

func (m *Mount) inUse() bool {
	if err := m.useChecker.Start(); err != nil {
		log.Fatalf("inUse: Start: %v", err)
	}
	if err := m.useChecker.Wait(); err != nil {
		return false
	}

	return true
}

func (m *Mount) mount() {
	go func() {
		m.killed = <-m.kill
		m.unmount()
	}()

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

			if m.killed {
				break
			}
		}
	}
}
