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

// SetSyncMode
// gSyncMode: true - wait for log message to be sent to ES before exiting (sync mode)
// gSyncMode: false - default, send log message to ES in goroutine and return immediately
func SetSyncMode(sync bool) {
	gSync = sync
}

// Printf is a wrapper around Printf(...) that supports logging and removes redacted data.
func Printf(format string, args ...interface{}) {
	// Actual logging to stdout & DB
	now := time.Now()
	msg := FilterRedacted(fmt.Sprintf("%s: "+format, append([]interface{}{ToYMDHMSDate(now)}, args...)...))
	_, err := fmt.Printf("%s", msg)
	if err != nil {
		log.Printf("Err: %s", err.Error())
	}
	if gLogger != nil {
		logf := func() {
			_, err := fmt.Printf(">>> %s", msg)
			_ = gLogger.Write(&logger.Log{
				Connector:     gLoggerConnector,
				Configuration: gLoggerConfiguration,
				Status:        gLoggerStatus,
				CreatedAt:     time.Now(),
				Message:       msg,
			})
			_, err = fmt.Printf("<<< %s", msg)
		}
		if gSync {
			logf()
			return
		}
		go logf()
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
