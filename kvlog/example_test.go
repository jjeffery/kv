package kvlog_test

import (
	"log"

	"github.com/jjeffery/kv/kvlog"
)

func Example() {
	// optionally setup log prefix and flags
	log.SetPrefix("test program: ")
	log.SetFlags(log.LstdFlags | log.LUTC)

	// attach kvlog for printing to stderr
	kvlog.Attach()

	log.Println("program started")
}
