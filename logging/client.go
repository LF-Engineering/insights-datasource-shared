package logging

import (
	"fmt"
	"time"

	"github.com/LF-Engineering/insights-datasource-shared/uuid"
	jsoniter "github.com/json-iterator/go"
)

const (
	logIndex   = "insights-task-logging"
	inProgress = "inprogress"
	failed     = "failed"
	done       = "done"
)

// ESLogProvider used in connecting to ES logging server
type ESLogProvider interface {
	CreateDocument(index, documentID string, body []byte) ([]byte, error)
	Get(index string, query map[string]interface{}, result interface{}) error
	UpdateDocument(index string, id string, body interface{}) ([]byte, error)
}

// LogProvider ...
type LogProvider struct {
	esClient    ESLogProvider
	environment string
}

// NewLogProvider ...
func NewLogProvider(esClient ESLogProvider, environment string) (*LogProvider, error) {
	logProvider := &LogProvider{
		esClient:    esClient,
		environment: environment,
	}

	return logProvider, nil
}

// StoreLog ...
func (s *LogProvider) StoreLog(log Log) error {
	if log.Datasource == "" || log.Endpoint == "" || log.CreatedAt.IsZero() {
		return fmt.Errorf("error: log datasource, endpoint and created at are all required")
	}
	if log.Status != inProgress && log.Status != failed && log.Status != done {
		return fmt.Errorf("error: log status must be one of [%s, %s, %s ]", inProgress, failed, done)
	}

	date := log.CreatedAt.Format(time.RFC3339)
	docID, err := uuid.Generate(log.Datasource, log.Endpoint, log.Status, date)
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

	return s.updateDocument(log, index, docID)
}

// PullLogs ...
func (s *LogProvider) PullLogs(datasource string) ([]Log, error) {
	must := make([]map[string]interface{}, 0)
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"datasource": map[string]string{
				"value": datasource},
		},
	})
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"status": map[string]string{
				"value": inProgress},
		},
	})
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				//"must": must,
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

func (s *LogProvider) updateDocument(log Log, index string, docID string) error {
	doc := map[string]interface{}{
		"datasource": log.Datasource,
		"endpoint":   log.Endpoint,
		"created_at": log.CreatedAt,
		"status":     log.Status,
	}

	_, err := s.esClient.UpdateDocument(index, docID, doc)
	if err != nil {
		fmt.Println(index)
		return err
	}
	return nil
}
