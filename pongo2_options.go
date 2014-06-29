package pongo2

import (
	"fmt"
	"log"
	"os"
)

type pongo2Options struct {
	debug bool
}

var (
	options = pongo2Options{}
	logger  = log.New(os.Stdout, "[pongo2] ", log.LstdFlags)
)

func SetDebug(b bool) {
	options.debug = b
}

func logf(format string, items ...interface{}) {
	if options.debug {
		logger.Printf(format, items...)
	}
}

func Logf(sender string, format string, items ...interface{}) {
	if options.debug {
		logger.Printf(fmt.Sprintln("[%s] %s", sender, format), items...)
	}
}
