package aws

import (
	"encoding/json"
	"fmt"
	"github.com/LF-Engineering/insights-datasource-shared/http"
	"os"
	"time"
)

// GetContainerARN ...
func GetContainerARN() (string, error) {
	httpClient := http.NewClientProvider(60*time.Second, true)
	statusCode, res, err := httpClient.Request(fmt.Sprintf("%s/task", os.Getenv("ECS_CONTAINER_METADATA_URI_V4")), "GET", nil, nil, nil)
	if err != nil {
		return "", err
	}
	if statusCode > 200 {
		return "", fmt.Errorf("getContainerMetadata error: status code is %d", statusCode)
	}

	metaResponse := ContainerMetadata{}
	err = json.Unmarshal(res, &metaResponse)
	if err != nil {
		return "", err
	}

	return metaResponse.TaskARN, nil
}

// ContainerMetadata ...
type ContainerMetadata struct {
	TaskARN string `json:"TaskARN"`
}
