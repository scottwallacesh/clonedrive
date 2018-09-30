package rclone

import (
	"clonedrive/lib"
	"fmt"
	"os/exec"
	"time"
)

// Move struct
type Move struct {
	move         []string
	source       string
	rcloneRemote string
	SleepTime    time.Duration
	schedule     string
	Kill         chan bool
	Killed       bool
}

// SetSchedule method
func (m *Move) SetSchedule(schedule string) {
	m.move = append(m.move, "--bwlimit="+schedule)
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
	go func(m *Move) {
		fmt.Printf("Waiting for mover kill signal\n")
		m.Killed = <-m.Kill
		fmt.Printf("Recieved mover kill signal\n")
	}(m)

	command := exec.Command(string(m.move[0]), m.move[1:]...)
	command.Dir = m.source
	command.Run()
}

// Mover constructor
func Mover(src string, dest string) *Move {
	newMove := Move{source: src, rcloneRemote: dest + ":"}
	newMove.move = []string{
		RclonePath,
		"move",
		".",
		newMove.rcloneRemote,
		"--exclude=.unionfs",
	}

	// Default sleepTime of 6 hours
	newMove.SleepTime = 6 * time.Hour

	newMove.Killed = false

	return &newMove
}
