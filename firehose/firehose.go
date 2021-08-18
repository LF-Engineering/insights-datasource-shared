package firehose

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"log"
	"os"
)

const (
	region        = "AWS_REGION"
	defaultRegion = "us-east-1"
)

type Config struct {
	Endpoint string
	Region   string
}

// ClientProvider for kinesis firehose
type ClientProvider struct {
	firehose *firehose.Firehose
	region   string
	endPoint string
}

// NewClientProvider initiate new client provider
func NewClientProvider() *ClientProvider {
	c := &ClientProvider{}
	c.region = os.Getenv(region)
	if c.region == "" {
		log.Printf("No AWS Region found for env var AWS_REGION. setting defaultRegion=%s \n", defaultRegion)
		c.region = defaultRegion
	}

	if os.Getenv("LOCALSTACK_HOSTNAME") != "" {
		c.endPoint = os.Getenv("LOCALSTACK_HOSTNAME")
	}
	sess := session.Must(session.NewSession())
	c.firehose = firehose.New(sess, aws.NewConfig().WithRegion(c.region).WithEndpoint(c.endPoint))

	return c
}

func LoadConfig() *Config {
	r := os.Getenv(region)
	if r == "" {
		log.Printf("No AWS Region found for env var AWS_REGION. setting defaultRegion=%s \n", defaultRegion)
		r = defaultRegion
	}
	cfg := Config{
		Region: r,
	}

	if os.Getenv("LOCALSTACK_HOSTNAME") != "" {
		cfg.Endpoint = os.Getenv("LOCALSTACK_HOSTNAME")
	}
	return &cfg
}

func (c *ClientProvider) PutRecordBatch(channel string, records []interface{}) (map[string]error, error) {
	inputs := make([]*firehose.Record, 0)
	for _, r := range records {
		b, err := json.Marshal(r)
		if err != nil {
			return map[string]error{}, err
		}
		records = append(records, &firehose.Record{Data: b})
	}

	params := &firehose.PutRecordBatchInput{
		DeliveryStreamName: aws.String(channel),
		Records:            inputs,
	}
	recordBatch, err := c.firehose.PutRecordBatch(params)
	if err != nil {
		return map[string]error{}, err
	}

	results := make(map[string]error)
	for _, res := range recordBatch.RequestResponses {
		if res.RecordId != nil {
			if res.ErrorMessage != nil && *res.ErrorMessage != "" {
				results[*res.RecordId] = errors.New(*res.ErrorMessage)
			} else {
				results[*res.RecordId] = nil
			}
		}

	}
	return results, nil
}

// PutRecord ...
func (c *ClientProvider) PutRecord(channel string, record interface{}) (string, error) {
	b, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	params := &firehose.PutRecordInput{
		DeliveryStreamName: aws.String(channel),
		Record:            &firehose.Record{Data: b},
	}
	res, err := c.firehose.PutRecord(params)
	if err != nil {
		return "", err
	}

	return *res.RecordId, nil
}
