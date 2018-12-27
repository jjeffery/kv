package main

import (
	"log"
	"os"

	"github.com/jjeffery/kv"
	"github.com/jjeffery/kv/kvlog"
)

func main() {
	kvlog.Attach()

	log.Println("debug: program started", kv.With("pid", os.Getpid()))
	defer log.Println("debug: program stopped", kv.With("pid", os.Getpid()))

	text := "info: lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod" +
		" tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis" +
		" nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis" +
		" aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat" +
		" nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui" +
		" officia deserunt mollit anim id est laborum."

	log.Println(text, kv.With(
		"mollit", "letraset",
		"occaecat", "voluptate",
		"cillum", "consectetur adipiscing",
		"price", "\u20ac250.00",
	))

	log.SetFlags(log.Flags() | log.Lshortfile)
	kvlog.Attach()
	log.Println("info: message after logger changed")
	for i := 1; i <= 5; i++ {
		log.Println("info: another message", kv.With("count", i))
	}
}
