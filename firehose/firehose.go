package firehose

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
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

// CreateDeliveryStream creating firehose delivery stream channel
// You must provide channel name as required parameter
// If channel created successfully it will return nil else it will return error
func (c *ClientProvider) CreateDeliveryStream(channel string)  error {
   deliveryType := firehose.DeliveryStreamTypeDirectPut
	params := &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(channel),
		DeliveryStreamType : &deliveryType,
	}
	req, _ := c.firehose.CreateDeliveryStreamRequest(params)
	 err:=  req.Send()
	if err !=nil{
		return err
	}
	return nil
}

// PutRecordBatch is operation for Amazon Kinesis Firehose
// Writes multiple data records into a delivery stream in a single call, which
// can achieve higher throughput per producer than when writing single records.
//
// Each PutRecordBatch request supports up to 500 records. Each record in the
// request can be as large as 1,000 KB (before 64-bit encoding), up to a limit
// of 4 MB for the entire request.
//
// You must specify the name of the delivery stream and the data record when
// using PutRecord. The data record consists of a data blob that can be up to
// 1,000 KB in size.
//
// The PutRecordBatch response includes a map of failed records.
// Even if the PutRecordBatch call succeeds
//
// Data records sent to Kinesis Data Firehose are stored for 24 hours from the
// time they are added to a delivery stream as it attempts to send the records
// to the destination. If the destination is unreachable for more than 24 hours,
// the data is no longer available.
//
// Don't concatenate two or more base64 strings to form the data fields of your
// records. Instead, concatenate the raw data, then perform base64 encoding.
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

// PutRecord is operation for Amazon Kinesis Firehose.
// Writes a single data record into an Amazon Kinesis Data Firehose delivery
// stream.
//
// By default, each delivery stream can take in up to 2,000 transactions per
// second, 5,000 records per second, or 5 MB per second.
//
// You must specify the name of the delivery stream and the data record when
// using PutRecord. The data record consists of a data blob that can be up to
// 1,000 KB in size, and any kind of data. You must specify the name of the delivery stream and the data record when
// using PutRecord. The data record consists of a data blob that can be up to
// 1,000 KB in size, and any kind of data.
//
// Kinesis Data Firehose buffers records before delivering them to the destination.
// To disambiguate the data blobs at the destination, a common solution is to
// use delimiters in the data, such as a newline (\n) or some other character
// unique within the data. This allows the consumer application to parse individual
// data items when reading the data from the destination.
//
// The PutRecord operation returns a RecordId, which is a unique string assigned
// to each record.
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