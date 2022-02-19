package cache

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

const Path = "cache/%s/%s"

type Manager struct {
	S3Provider  S3Provider
	Environment string
	Connector   string
}

func NewManager(s3Provider S3Provider) *Manager {
	return &Manager{
		S3Provider: s3Provider,
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

func (m *Manager) CreateCache(data []map[string]interface{}) error {
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
