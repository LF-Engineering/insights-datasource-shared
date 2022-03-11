package ingestjob

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LF-Engineering/insights-datasource-shared/uuid"
	jsoniter "github.com/json-iterator/go"
)

const (
	logIndex   = "insights-connector"
	tasksIndex = "insights-tasks"
	InProgress = "inprogress"
	Failed     = "failed"
	Done       = "done"
	Internal   = "internal"
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
	if log.Status != InProgress && log.Status != Failed && log.Status != Done && log.Status != Internal {
		return fmt.Errorf("error: log status must be one of [%s, %s, %s, %s]", InProgress, Failed, Done, Internal)
	}

	docID, err := generateID(log)
	if err != nil {
		return err
	}

	if log.UpdatedAt.IsZero() {
		log.UpdatedAt = log.CreatedAt
	}
	b, err := jsoniter.Marshal(log)
	if err != nil {
		return err
	}

	index := fmt.Sprintf("%s-%s-log-%s", logIndex, log.Connector, s.environment)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"_id": map[string]string{
					"value": docID},
			},
		},
	}

	var res TopHits
	err = s.esClient.Get(fmt.Sprintf("%s-%s-log-%s", logIndex, log.Connector, s.environment), query, &res)
	if err != nil || len(res.Hits.Hits) == 0 {
		_, err := s.esClient.CreateDocument(index, docID, b)
		return err
	}

	return s.updateDocument(*log, index, docID)
}

// Read ...
func (s *Logger) Read(connector string, status string) ([]Log, error) {
	if status != InProgress && status != Failed && status != Done && status != Internal {
		return []Log{}, fmt.Errorf("error: log status must be one of [%s, %s, %s, %s]", InProgress, Failed, Done, Internal)
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
	err := s.esClient.Get(fmt.Sprintf("%s-%s-log-%s", logIndex, connector, s.environment), query, &res)
	if err != nil {
		return logs, err
	}

	for _, l := range res.Hits.Hits {
		logs = append(logs, l.Source)
	}

	return logs, nil
}

func (s *Logger) Count(connector string, status string) (int, error) {
	if status != InProgress && status != Failed && status != Done && status != Internal {
		return 0, fmt.Errorf("error: log status must be one of [%s, %s, %s, %s]", InProgress, Failed, Done, Internal)
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
	return s.esClient.Count(fmt.Sprintf("%s-%s-log-%s", logIndex, connector, s.environment), query)
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

// Filter connector logs based on status, configuration and creation date
func (s *Logger) Filter(log *Log) ([]Log, error) {
	if log.Connector == "" {
		return []Log{}, fmt.Errorf("error: log connector is required")
	}
	if log.Status != InProgress && log.Status != Failed && log.Status != Done && log.Status != Internal {
		return []Log{}, fmt.Errorf("error: log status must be one of [%s, %s, %s, %s]", InProgress, Failed, Done, Internal)
	}

	must := CreateMustTerms(log)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
	}

	var res TopHits
	logs := make([]Log, 0)
	err := s.esClient.Get(fmt.Sprintf("%s-%s-log-%s", logIndex, log.Connector, s.environment), query, &res)
	if err != nil {
		return logs, err
	}

	for _, l := range res.Hits.Hits {
		logs = append(logs, l.Source)
	}

	return logs, nil
}

func CreateMustTerms(log *Log) []map[string]interface{} {
	must := make([]map[string]interface{}, 0)

	if log.Status != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"status": map[string]string{
					"value": log.Status},
			},
		})
	}

	if len(log.Configuration) != 0 {
		for _, conf := range log.Configuration {
			for k, v := range conf {
				val := strings.ReplaceAll(v, "/", "\\/")
				must = append(must, map[string]interface{}{
					"query_string": map[string]interface{}{
						"default_field": fmt.Sprintf("configuration.%s", k),
						"query":         val,
					},
				})
			}
		}
	}

	if log.From != nil {
		from := log.From.Format(time.RFC3339)
		to := "now/d"
		if log.To != nil {
			to = log.To.Format(time.RFC3339)
		}
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{
				"created_at": map[string]string{
					"gte": from,
					"lte": to},
			},
		})
	}

	return must
}

func generateID(log *Log) (string, error) {
	date := log.CreatedAt.Format(time.RFC3339)
	configs, err := json.Marshal(log.Configuration)
	if err != nil {
		return "", err
	}
	docID, err := uuid.Generate(log.Connector, string(configs), date)
	if err != nil {
		return "", err
	}
	return docID, nil
}

// WriteTask ...
func (s *Logger) WriteTask(log *TaskLog) error {
	if log.Connector == "" || len(log.Configuration) == 0 || log.CreatedAt.IsZero() {
		return fmt.Errorf("error: log connector, configuration and created at are all required")
	}

	docID, err := generateTaskID(log)
	if err != nil {
		return err
	}

	b, err := jsoniter.Marshal(log)
	if err != nil {
		return err
	}

	index := fmt.Sprintf("%s-%s-log-%s", tasksIndex, log.Connector, s.environment)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"_id": map[string]string{
					"value": docID},
			},
		},
	}

	var res TopHits
	err = s.esClient.Get(fmt.Sprintf("%s-%s-log-%s", logIndex, log.Connector, s.environment), query, &res)
	if err != nil || len(res.Hits.Hits) == 0 {
		_, err := s.esClient.CreateDocument(index, docID, b)
		return err
	}
	return err
}

func generateTaskID(log *TaskLog) (string, error) {
	date := log.CreatedAt.Format(time.RFC3339)
	configs, err := json.Marshal(log.Configuration)
	if err != nil {
		return "", err
	}
	docID, err := uuid.Generate(log.Connector, string(configs), date)
	if err != nil {
		return "", err
	}
	return docID, nil
}
