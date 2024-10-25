package logger

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	loggerLevel  = flag.String("logger.level", "INFO", "Minimum level to log. Possible values: INFO, WARN, ERROR, FATAL, PANIC.")
	loggerOutput = flag.String("logger.output", "stderr", "Output for the logs. Supported values: stderr, stdout.")
	mu           sync.Mutex
	output       io.Writer = os.Stderr
	timezone               = time.UTC
)

const (
	programRootPath       = "/webCrawler/"
	programRootPathLength = len(programRootPath)
)

func Initialization() {
	validateLoggerLevel()
	setLoggerOutput()
}

func validateLoggerLevel() {
	switch *loggerLevel {
	case "INFO", "WARN", "ERROR", "FATAL", "PANIC":
	default:
		// We cannot use logger.Panicf here, since the logger hasn't been initialized yet.
		panic(fmt.Errorf("FATAL: unsupported `-loggerLevel` value: %q; supported values are: INFO, WARN, ERROR, FATAL, PANIC", *loggerLevel))
	}
}

func setLoggerOutput() {
	switch *loggerOutput {
	case "stderr":
		output = os.Stderr
	case "stdout":
		output = os.Stdout
	default:
		panic(fmt.Errorf("FATAL: unsupported `loggerOutput` value: %q; supported values are: stderr, stdout", *loggerOutput))
	}
}

// Infof logs info message.
func Infof(format string, args ...interface{}) {
	logLevel("INFO", format, args)
}

// Warnf logs warn message.
func Warnf(format string, args ...interface{}) {
	logLevel("WARN", format, args)
}

// Errorf logs error message.
func Errorf(format string, args ...interface{}) {
	logLevel("ERROR", format, args)
}

// Fatalf logs fatal message and terminates the app.
func Fatalf(format string, args ...interface{}) {
	logLevel("FATAL", format, args)
}

// Panicf logs panic message and panics.
func Panicf(format string, args ...interface{}) {
	logLevel("PANIC", format, args)
}

func logLevel(level, format string, args []interface{}) {
	if shouldSkipLog(level) {
		return
	}
	var msg string = formatLogMessage(format, args)
	logMessage(level, msg, 3)
}

func formatLogMessage(format string, args []interface{}) string {
	// Might add a feature to limit the length of message in the future.
	return fmt.Sprintf(format, args...)
}

func logMessage(level, msg string, skipframes int) {
	var timestamp string = time.Now().In(timezone).Format("2006-01-02T15:04:05.000Z0700")

	_, file, line, ok := runtime.Caller(skipframes)
	if !ok {
		file = "???"
		line = 0
	}

	if n := strings.Index(file, programRootPath); n >= 0 {
		// Strip /secretRotation/ prefix
		file = file[n+programRootPathLength:]
	}
	location := fmt.Sprintf("%s:%d", file, line)

	// Might add a suppression feature here in the future.

	for len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	var logMsg string = fmt.Sprintf("%s\t%s\t%s\t%s\n", timestamp, level, location, msg)
	// Serialize writes to log.
	mu.Lock()
	fmt.Fprint(output, logMsg)
	mu.Unlock()

	switch level {
	case "PANIC":
		panic(errors.New(msg))
	case "FATAL":
		os.Exit(-1)
	}
}

func shouldSkipLog(level string) bool {
	switch *loggerLevel {
	case "WARN":
		switch level {
		case "WARN", "ERROR", "FATAL", "PANIC":
			return false
		default:
			return true
		}
	case "ERROR":
		switch level {
		case "ERROR", "FATAL", "PANIC":
			return false
		default:
			return true
		}
	case "FATAL":
		switch level {
		case "FATAL", "PANIC":
			return false
		default:
			return true
		}
	case "PANIC":
		return level != "PANIC"
	default:
		return false
	}
}
