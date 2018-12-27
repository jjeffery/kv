// Package kvlog handles messages printed by the `log` package
// in the Go standard library. It parses the message for key/value
// pairs and formats the message on terminals with line wrapping
// and color display.
package kvlog

import (
	"io"

	"github.com/jjeffery/kv/internal/terminal"
)

// IsTerminal returns true if the writer is a terminal.
func IsTerminal(writer io.Writer) bool {
	if fd, ok := fileDescriptor(writer); ok {
		return terminal.IsTerminal(fd)
	}
	return false
}
