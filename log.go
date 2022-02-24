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

// Printf is a wrapper around Printf(...) that supports logging and removes redacted data.
func Printf(format string, args ...interface{}) {
	// TODO: FIXME: wrapper log function that also pushes everything to log elasticsearch cluster (cc @Fayaz)
	// Actual logging to stdout & DB
	now := time.Now()
	msg := FilterRedacted(fmt.Sprintf("%s: "+format, append([]interface{}{ToYMDHMSDate(now)}, args...)...))
	_, err := fmt.Printf("%s", msg)
	if err != nil {
		log.Printf("Err: %s", err.Error())
	}
	if gLogger != nil {
		go func() {
			_ = gLogger.Write(&logger.Log{
				Connector:     gLoggerConnector,
				Configuration: gLoggerConfiguration,
				Status:        gLoggerStatus,
				CreatedAt:     time.Now(),
				Message:       msg,
			})
		}()
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
