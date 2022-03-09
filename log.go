package ds

import (
	"fmt"
	"log"
	"time"

	logger "github.com/LF-Engineering/insights-datasource-shared/ingestjob"
)

var (
	gLogger              *logger.Logger
	gLoggerConnector     string
	gLoggerStatus        string
	gLoggerConfiguration []map[string]string
	gSync                bool
	gConsoleAfterES      bool
)

// AddLogger - adds logger
func AddLogger(logger *logger.Logger, connector, status string, configuration []map[string]string) {
	if logger != nil {
		gLogger = logger
		gLoggerConnector = connector
		gLoggerConfiguration = configuration
		gLoggerStatus = status
	}
}

// SetSyncMode - sets sync/async ES loging mode
// sync -> gSyncMode: true - wait for log message to be sent to ES before exiting (sync mode)
// sync -> gSyncMode: false - default, send log message to ES in goroutine and return immediately
// consoleAfterES -> gConsoleAfterES - will log on console after logged to ES
func SetSyncMode(sync, consoleAfterES bool) {
	gSync = sync
	gConsoleAfterES = consoleAfterES
}

// Printf is a wrapper around Printf(...) that supports logging and removes redacted data.
func Printf(format string, args ...interface{}) {
	// Actual logging to stdout & DB
	now := time.Now()
	msg := FilterRedacted(fmt.Sprintf("%s: "+format, append([]interface{}{ToYMDHMSDate(now)}, args...)...))
	logConsole := func() {
		_, err := fmt.Printf("%s", msg)
		if err != nil {
			log.Printf("Error (log to console): %s", err.Error())
		}
	}
	if !gConsoleAfterES || gLogger == nil {
		logConsole()
	}
	if gLogger != nil {
		logES := func() {
			err := gLogger.Write(&logger.Log{
				Connector:     gLoggerConnector,
				Configuration: gLoggerConfiguration,
				Status:        gLoggerStatus,
				CreatedAt:     time.Now(),
				Message:       msg,
			})
			if err != nil {
				log.Printf("Error (log to ES): %s", err.Error())
			}
			if gConsoleAfterES {
				logConsole()
			}
		}
		if gSync {
			logES()
			return
		}
		go logES()
	}
}

// PrintfNoRedacted is a wrapper around Printf(...) that supports logging and don't removes redacted data
func PrintfNoRedacted(format string, args ...interface{}) {
	// Actual logging to stdout & DB
	now := time.Now()
	msg := fmt.Sprintf("%s: "+format, append([]interface{}{ToYMDHMSDate(now)}, args...)...)
	_, err := fmt.Printf("%s", msg)
	if err != nil {
		log.Printf("Err: %s", err.Error())
	}
}
