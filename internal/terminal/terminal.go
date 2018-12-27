// This code was adapted from the following package:
//
//  golang.org/x/crypt/ssh/terminal.
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux,!appengine netbsd openbsd

// Package terminal provides support functions for dealing with terminals.
package terminal

import (
	"golang.org/x/sys/unix"
)

// State contains the state of a terminal.
type State struct {
	termios unix.Termios
}

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	return err == nil
}

// EnableVirtualTerminalProcessing configures the terminal to accept ANSI
// sequences. This is a no-op for all operating systems other than Windows.
func EnableVirtualTerminalProcessing(fd int) error {
	return nil
}

// GetSize returns the dimensions of the given terminal.
func GetSize(fd int) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}
	return int(ws.Col), int(ws.Row), nil
}
