package main

import (
	"os/exec"
	"syscall"
	"time"
)

// Mover struct
type Mover struct {
	mover        exec.Cmd
	source       string
	rcloneRemote string
	sleepTime    time.Duration
	schedule     string
	kill         chan bool
	killed       bool
}

func (m *Mover) setSchedule(schedule string) {
	m.mover.Args = append(m.mover.Args, "--bwlimit="+schedule)
}

func (m *Mover) setSleepTime(sleepTime string) bool {
	if seconds, err := convertTime(sleepTime); err == nil {
		m.sleepTime = time.Duration(seconds) * time.Second
		return true
	}

	return false
}

func (m *Mover) move() {
	go func() {
		// Wait for the kill signal
		m.killed = <-m.kill

		// Stop the rclone mover
		m.mover.Process.Signal(syscall.SIGINT)
	}()

	for {
		m.mover.Run()
		if m.killed {
			break
		}
		time.Sleep(m.sleepTime)
	}
}

// RcloneMover constructor
func RcloneMover(src string, dest string) *Mover {
	newMover := Mover{source: src, rcloneRemote: dest + ":"}
	newMover.mover = *exec.Command(rclonePath,
		"move",
		".",
		newMover.rcloneRemote,
		"--exclude=.unionfs")
	newMover.mover.Dir = newMover.source

	// Default sleepTime of 6 hours
	newMover.sleepTime = 6 * time.Hour

	newMover.killed = false

	return &newMover
}
