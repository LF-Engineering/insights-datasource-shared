package cache

import (
	"encoding/json"
	"fmt"
	s3util "github.com/LF-Engineering/insights-datasource-shared/aws/s3"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"time"
)

const (
	Path         = "cache/%s/%s/%s"
	LastSyncFile = "0000-last-sync"
	Bucket       = "insights-v2-cache-%s"
	Region       = "us-east-2"
)

type Manager struct {
	s3Manager S3Manager
	connector string
	endpoint  string
}

func NewManager(connector string, environment string) *Manager {
	s3Manager := s3util.NewManager(fmt.Sprintf(Bucket, environment), Region)
	return &Manager{
		s3Manager: s3Manager,
		connector: connector,
	}
}

// S3Manager used in connecting to s3
type S3Manager interface {
	Save(payload []byte) error
	SaveWithKey(payload []byte, key string) error
	GetKeys() ([]string, error)
	Get(key string) ([]byte, error)
	Delete(key string) error
}

// IsKeyCreated check if the key already exists
func (m *Manager) IsKeyCreated(endpoint string, id string) (bool, error) {
	key := fmt.Sprintf(Path, m.connector, endpoint, id)
	_, err := m.s3Manager.Get(key)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

// Create new cache record
func (m *Manager) Create(endpoint string, data []map[string]interface{}) error {
	for _, v := range data {
		id, ok := v["id"]
		if !ok {
			return fmt.Errorf("error getting id")
		}
		key := fmt.Sprintf(Path, m.connector, endpoint, id)
		b, err := json.Marshal(v["data"])
		if err != nil {
			return err
		}

		err = m.s3Manager.SaveWithKey(b, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetLastSync get connector sync date, if it is not exist return epoch date
func (m *Manager) GetLastSync(endpoint string) (time.Time, error) {
	key := fmt.Sprintf(Path, m.connector, endpoint, LastSyncFile)
	from, err := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	if err != nil {
		return from, err
	}
	d, err := m.s3Manager.Get(key)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return from, m.SetLastSync(endpoint, from)
			default:
				return from, err
			}
		}
		return from, err
	}
	err = json.Unmarshal(d, &from)
	if err != nil {
		return from, err
	}

	return from, nil
}

// SetLastSync update connector last sync date
func (m *Manager) SetLastSync(endpoint string, lastSync time.Time) error {
	key := fmt.Sprintf(Path, m.connector, endpoint, LastSyncFile)
	b, err := json.Marshal(lastSync)
	if err != nil {
		return err
	}
	err = m.s3Manager.SaveWithKey(b, key)
	if err != nil {
		return err
	}
	return nil
}

// GetFileByKey get file by key
func (m *Manager) GetFileByKey(endpoint string, id string) ([]byte, error) {
	key := fmt.Sprintf(Path, m.connector, endpoint, id)
	data, err := m.s3Manager.Get(key)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return data, nil
			default:
				return data, err
			}
		}
		return data, err
	}
	return data, nil
}

// UpdateFileByKey get file by key
func (m *Manager) UpdateFileByKey(endpoint string, id string, data []byte) error {
	key := fmt.Sprintf(Path, m.connector, endpoint, id)
	err := m.s3Manager.SaveWithKey(data, key)
	if err != nil {
		return err
	}
	return nil
}
