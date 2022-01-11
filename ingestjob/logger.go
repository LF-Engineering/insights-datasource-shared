package ingestjob

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LF-Engineering/insights-datasource-shared/uuid"
	jsoniter "github.com/json-iterator/go"
)

const (
	logIndex   = "insights-job-logging"
	inProgress = "inprogress"
	failed     = "failed"
	done       = "done"
)

// ESLogProvider used in connecting to ES logging server
type ESLogProvider interface {
	CreateDocument(index, documentID string, body []byte) ([]byte, error)
	Get(index string, query map[string]interface{}, result interface{}) error
	UpdateDocument(index string, id string, body interface{}) ([]byte, error)
	Count(index string, query map[string]interface{}) (int, error)
}

// Logger ...
type Logger struct {
	esClient    ESLogProvider
	environment string
}

// NewLogger ...
func NewLogger(esClient ESLogProvider, environment string) (*Logger, error) {
	logProvider := &Logger{
		esClient:    esClient,
		environment: environment,
	}

	return logProvider, nil
}

// Write ...
func (s *Logger) Write(log *Log) error {
	if log.Connector == "" || len(log.Configuration) == 0 || log.CreatedAt.IsZero() {
		return fmt.Errorf("error: log connector, configuration and created at are all required")
	}
	if log.Status != inProgress && log.Status != failed && log.Status != done {
		return fmt.Errorf("error: log status must be one of [%s, %s, %s ]", inProgress, failed, done)
	}

	date := log.CreatedAt.Format(time.RFC3339)
	configs, err := json.Marshal(log.Configuration)
	if err != nil {
		return err
	}
	docID, err := uuid.Generate(log.Connector, string(configs), date)
	if err != nil {
		return err
	}

	b, err := jsoniter.Marshal(log)
	if err != nil {
		return err
	}

	index := fmt.Sprintf("%s-%s", logIndex, s.environment)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"_id": map[string]string{
					"value": docID},
			},
		},
	}

	var res TopHits
	err = s.esClient.Get(fmt.Sprintf("%s-%s", logIndex, s.environment), query, &res)
	if err != nil || len(res.Hits.Hits) == 0 {
		_, err := s.esClient.CreateDocument(index, docID, b)
		return err
	}

	return s.updateDocument(*log, index, docID)
}

// Read ...
func (s *Logger) Read(connector string, status string) ([]Log, error) {
	if status != inProgress && status != failed && status != done {
		return []Log{}, fmt.Errorf("error: log status must be one of [%s, %s, %s ]", inProgress, failed, done)
	}

	must := make([]map[string]interface{}, 0)
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"connector": map[string]string{
				"value": connector},
		},
	})
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"status": map[string]string{
				"value": status},
		},
	})
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
	}

	var res TopHits
	logs := make([]Log, 0)
	err := s.esClient.Get(fmt.Sprintf("%s-%s", logIndex, s.environment), query, &res)
	if err != nil {
		return logs, err
	}

	for _, l := range res.Hits.Hits {
		logs = append(logs, l.Source)
	}

	return logs, nil
}

func (s *Logger) Count(connector string, status string) (int, error) {
	if status != inProgress && status != failed && status != done {
		return 0, fmt.Errorf("error: log status must be one of [%s, %s, %s ]", inProgress, failed, done)
	}

	must := make([]map[string]interface{}, 0)
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"connector": map[string]string{
				"value": connector},
		},
	})
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"status": map[string]string{
				"value": status},
		},
	})
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
	}

	return s.esClient.Count(fmt.Sprintf("%s-%s", logIndex, s.environment), query)
}

func (s *Logger) updateDocument(log Log, index string, docID string) error {
	doc := map[string]interface{}{
		"connector":     log.Connector,
		"configuration": log.Configuration,
		"updated_at":    time.Now().UTC(),
		"status":        log.Status,
	}

	_, err := s.esClient.UpdateDocument(index, docID, doc)
	if err != nil {
		return err
	}
	return nil
}
