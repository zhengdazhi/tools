package logutil

import (
	"log"
)

var debugMode bool

// SetDebug enables or disables debug mode
func SetDebug(debug bool) {
	debugMode = debug
}

// LogDebug logs debug messages if debug mode is enabled
func LogDebug(format string, v ...interface{}) {
	if debugMode {
		log.Printf(format, v...)
	}
}
