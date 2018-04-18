package mounter

import (
	"log"
	"os/exec"
	"time"
)

// Mount struct
type Mount struct {
	Mounter    exec.Cmd
	unmounter  exec.Cmd
	useChecker exec.Cmd
	source     string
	mountPoint string
	ready      chan bool
	Overlay    bool
	Kill       chan bool
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

// Mount method
func (m *Mount) Mount() {
	go func() {
		m.killed = <-m.Kill
		m.unmount()
	}()

	for {
		// Wait for a ready signal on overlay mounts
		if m.Overlay == true {
			// Any signal will do
			<-m.ready
		}

		// Make sure nothing's mounted
		m.unmount()

		// This is a new mount session
		m.killed = false

		// If not in use
		if !m.inUse() {
			// Run the command, sleep briefly, signal the overlay and wait
			if err := m.Mounter.Start(); err != nil {
				log.Fatalf("mount: Start: %v", err)
			}

			time.Sleep(3 * time.Second)

			if m.Overlay == false {
				m.ready <- true
			}

			m.Mounter.Wait()

			if m.killed {
				break
			}
		}
	}
}
