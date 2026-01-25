package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile   *os.File
	logMutex  sync.Mutex
	logInited bool
)

// initDebugLog initializes the debug log file
func initDebugLog() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logInited {
		return nil
	}

	logPath := filepath.Join("/tmp", "petitorium_debug.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logFile = f
	logInited = true
	return nil
}

// debugLog writes a debug message to the log file
func debugLog(format string, args ...interface{}) {
	if !logInited {
		if err := initDebugLog(); err != nil {
			return
		}
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	logFile.WriteString(fmt.Sprintf("%s %s\n", timestamp, msg))
	logFile.Sync()
}
