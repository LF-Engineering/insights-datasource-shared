package cache

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"time"
)

const (
	Path         = "cache/%s/%s"
	LastSyncFile = "0000-last-sync"
)

type Manager struct {
	S3Provider S3Provider
	Connector  string
}

func NewManager(s3Provider S3Provider, connector string) *Manager {
	return &Manager{
		S3Provider: s3Provider,
		Connector:  connector,
	}
}

// S3Provider used in connecting to s3
type S3Provider interface {
	Save(payload []byte) error
	SaveWithKey(payload []byte, key string) error
	GetKeys() ([]string, error)
	Get(key string) ([]byte, error)
	Delete(key string) error
}

// IsKeyCreated check if the key already exists
func (m *Manager) IsKeyCreated(id string) (bool, error) {
	key := fmt.Sprintf(Path, m.Connector, id)
	_, err := m.S3Provider.Get(key)
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
func (m *Manager) Create(data []map[string]interface{}) error {
	for _, v := range data {
		id, ok := v["id"]
		if !ok {
			return fmt.Errorf("error getting id")
		}
		key := fmt.Sprintf(Path, m.Connector, id)
		b, err := json.Marshal(v["data"])
		if err != nil {
			return err
		}

		err = m.S3Provider.SaveWithKey(b, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetLastSync get connector sync date, if it is not exist return epoch date
func (m *Manager) GetLastSync() (time.Time, error) {
	key := fmt.Sprintf(Path, m.Connector, LastSyncFile)
	from, err := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	if err != nil {
		return from, err
	}
	d, err := m.S3Provider.Get(key)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return from, m.SetLastSync(from)
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
func (m *Manager) SetLastSync(lastSync time.Time) error {
	key := fmt.Sprintf(Path, m.Connector, LastSyncFile)
	b, err := json.Marshal(lastSync)
	if err != nil {
		return err
	}
	err = m.S3Provider.SaveWithKey(b, key)
	if err != nil {
		return err
	}
	return nil
}
