package debug

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	enabled bool
	logger  *log.Logger
	logFile *os.File
)

func Init(enable bool, logPath string) error {
	enabled = enable

	if !enabled {
		return nil
	}

	if logPath == "" {
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			dataDir = filepath.Join(home, ".local", "share")
		}
		logPath = filepath.Join(dataDir, "memento", "debug.log")
	}

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	logger = log.New(logFile, "", 0)
	Log("Debug logging initialized at %s", logPath)

	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func Log(format string, args ...interface{}) {
	if !enabled || logger == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logger.Printf("[%s] %s", timestamp, message)
}

func LogKeyEvent(key string) {
	Log("KEY: %s", key)
}

func LogAPIRequest(method, path string) {
	Log("API REQUEST: %s %s", method, path)
}

func LogAPIResponse(method, path string, status int, duration time.Duration) {
	Log("API RESPONSE: %s %s -> %d (%v)", method, path, status, duration)
}

func LogError(err error, context string) {
	Log("ERROR [%s]: %v", context, err)
}

func LogStateChange(from, to string) {
	Log("STATE: %s -> %s", from, to)
}

func Enabled() bool {
	return enabled
}
