package elasticgap

import (
	b64 "encoding/base64"
	"fmt"

	"github.com/LF-Engineering/insights-datasource-shared/elastic"
	jsoniter "github.com/json-iterator/go"
)

// HTTPClient used in connecting to remote http server
type HTTPClient interface {
	Request(url string, method string, header map[string]string, body []byte, params map[string]string) (statusCode int, resBody []byte, err error)
}

// Auth0Client ...
type Auth0Client interface {
	GetToken() (string, error)
}

// GapHandler ...
type GapHandler struct {
	gapURL      string
	httpClient  HTTPClient
	auth0Client Auth0Client
}

// NewGapHandler ...
func NewGapHandler(gapURL string, httpClient HTTPClient, auth0Client Auth0Client) *GapHandler {
	return &GapHandler{
		gapURL:      gapURL,
		httpClient:  httpClient,
		auth0Client: auth0Client,
	}
}

// Send unsaved data to data-gap handler
func (g *GapHandler) Send(data []elastic.BulkData) error {
	token, err := g.auth0Client.GetToken()
	if err != nil {
		return err
	}

	byteData, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}

	dataEnc := b64.StdEncoding.EncodeToString(byteData)
	gapBody := map[string]map[string]string{"index": {"content": dataEnc}}
	bData, err := jsoniter.Marshal(gapBody)
	if err != nil {
		return err
	}

	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", token)

	if g.gapURL != "" {
		_, _, err = g.httpClient.Request(g.gapURL, "POST", header, bData, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// HandleFailedData ...
func (g *GapHandler) HandleFailedData(data []elastic.BulkData, byteResponse []byte) (failedIndexes []elastic.BulkData, err error) {
	var esRes ElasticResponse
	err = jsoniter.Unmarshal(byteResponse, &esRes)
	if err != nil {
		return failedIndexes, err
	}

	// loop throw elastic response to get failed indexes
	for _, item := range esRes.Items {
		if item.Index.Status != 200 {
			var singleBulk elastic.BulkData
			// loop throw real data to get failed ones
			for _, el := range data {
				if el.ID == item.Index.ID {
					singleBulk = el
					break
				}
			}
			failedIndexes = append(failedIndexes, singleBulk)
		}
	}
	return failedIndexes, nil
}
