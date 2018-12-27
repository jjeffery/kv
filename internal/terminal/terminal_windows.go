// This code was adapted from the following package:
//
//  golang.org/x/crypt/ssh/terminal.
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

// Package terminal provides support functions for dealing with terminals.
package terminal

import (
	"golang.org/x/sys/windows"
)

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	var st uint32
	err := windows.GetConsoleMode(windows.Handle(fd), &st)
	return err == nil
}

// EnableVirtualTerminalProcessing configures the terminal to accept ANSI
// sequences. This is a no-op for all operating systems other than Windows.
func EnableVirtualTerminalProcessing(fd int) error {
	var st uint32
	err := windows.GetConsoleMode(windows.Handle(fd), &st)
	if err != nil {
		return err
	}
	if st&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == 0 {
		return windows.SetConsoleMode(windows.Handle(fd), st|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
	return nil
}

// GetSize returns the dimensions of the given terminal.
func GetSize(fd int) (width, height int, err error) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(fd), &info); err != nil {
		return 0, 0, err
	}
	return int(info.Size.X), int(info.Size.Y), nil
}
