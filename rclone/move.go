package rclone

import (
	"../lib"
	"os/exec"
	"time"
)

// Move struct
type Move struct {
	move         exec.Cmd
	source       string
	rcloneRemote string
	SleepTime    time.Duration
	schedule     string
	Kill         chan bool
	Killed       bool
}

// SetSchedule method
func (m *Move) SetSchedule(schedule string) {
	m.move.Args = append(m.move.Args, "--bwlimit="+schedule)
}

// SetSleepTime method
func (m *Move) SetSleepTime(sleepTime string) bool {
	if seconds, err := lib.ConvertTime(sleepTime); err == nil {
		m.SleepTime = time.Duration(seconds) * time.Second
		return true
	}

	return false
}

// Run method
func (m *Move) Run() {
	m.move.Dir = m.source
	m.move.Run()
}

// Mover constructor
func Mover(src string, dest string) *Move {
	newMove := Move{source: src, rcloneRemote: dest + ":"}
	newMove.move = *exec.Command(RclonePath,
		"move",
		".",
		newMove.rcloneRemote,
		"--exclude=.unionfs")

	// Default sleepTime of 6 hours
	newMove.SleepTime = 6 * time.Hour

	newMove.Killed = false

	return &newMove
}
