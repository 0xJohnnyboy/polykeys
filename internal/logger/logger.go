package logger

import (
	"fmt"
	"sync"
)

var (
	debugEnabled bool
	mu           sync.RWMutex
)

// SetDebug enables or disables debug logging
func SetDebug(enabled bool) {
	mu.Lock()
	defer mu.Unlock()
	debugEnabled = enabled
}

// IsDebug returns whether debug logging is enabled
func IsDebug() bool {
	mu.RLock()
	defer mu.RUnlock()
	return debugEnabled
}

// Debug prints a debug message if debug logging is enabled
func Debug(format string, args ...interface{}) {
	if IsDebug() {
		fmt.Printf(format, args...)
	}
}

// Debugln prints a debug message with newline if debug logging is enabled
func Debugln(args ...interface{}) {
	if IsDebug() {
		fmt.Println(args...)
	}
}
