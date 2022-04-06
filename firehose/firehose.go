package firehose

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

const (
	region        = "AWS_REGION"
	defaultRegion = "us-east-1"
	maxChunkSize  = 1020000
)

//Config aws configuration
type Config struct {
	Endpoint string
	Region   string
}

// PutResponse per record
type PutResponse struct {
	RecordID string
	Error    error
}

// ClientProvider for kinesis firehose
type ClientProvider struct {
	firehose *firehose.Client
	region   string
	endPoint string
}

// NewClientProvider initiate new client provider
func NewClientProvider() (*ClientProvider, error) {
	c := &ClientProvider{}
	c.region = os.Getenv(region)
	if c.region == "" {
		log.Printf("No AWS Region found for env var AWS_REGION. setting defaultRegion=%s \n", defaultRegion)
		c.region = defaultRegion
	}

	if os.Getenv("LOCALSTACK_HOSTNAME") != "" {
		c.endPoint = os.Getenv("LOCALSTACK_HOSTNAME")
	}

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if c.endPoint != "" {
			return aws.Endpoint{
				URL:           fmt.Sprintf("http://%s:4566", c.endPoint),
				SigningRegion: c.region,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fall back to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(c.region),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		return nil, err
	}
	c.firehose = firehose.NewFromConfig(cfg)

	return c, nil
}

// CreateDeliveryStream creating firehose delivery stream channel
// You must provide channel name as required parameter
// If channel created successfully it will return nil else it will return error
func (c *ClientProvider) CreateDeliveryStream(channel string) error {
	params := &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(channel),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
	}
	_, err := c.firehose.CreateDeliveryStream(context.Background(), params)
	return err
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
func (c *ClientProvider) PutRecordBatch(channel string, records []interface{}) ([]*PutResponse, error) {
	ch := make(chan *chanPutResponse)

	recordSize, err := size(records)
	if err != nil {
		return []*PutResponse{}, err
	}

	if recordSize < maxChunkSize {
		result, err := c.send(channel, records)
		if err != nil {
			return []*PutResponse{}, err
		}
		return result, nil
	}

	chunks, err := spiltRecords(records)
	if err != nil {
		return []*PutResponse{}, err
	}
	for _, chunk := range chunks {
		chunk := chunk
		go func() {
			result, err := c.send(channel, chunk)
			if err != nil {
				ch <- &chanPutResponse{Error: err}
			}
			ch <- &chanPutResponse{Result: result}
		}()
	}

	var res []*PutResponse
	for i := 0; i < len(chunks); i++ {
		select {
		case r := <-ch:
			if r.Error != nil {
				return []*PutResponse{}, r.Error
			}
			res = append(res, r.Result...)
		}
	}

	return res, nil
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
func (c *ClientProvider) PutRecord(channel string, record interface{}) (*PutResponse, error) {
	b, err := json.Marshal(record)
	if err != nil {
		return &PutResponse{}, err
	}
	if len(b) > maxChunkSize {
		return &PutResponse{}, errors.New("record exceeded the limit of 1 mb")
	}

	params := &firehose.PutRecordInput{
		DeliveryStreamName: aws.String(channel),
		Record:             &types.Record{Data: b},
	}
	res, err := c.firehose.PutRecord(context.Background(), params)
	if err != nil {
		return &PutResponse{}, err
	}
	return &PutResponse{RecordID: *res.RecordId, Error: nil}, nil
}

func spiltRecords(records []interface{}) ([][]interface{}, error) {
	chunks := make([][]interface{}, 0)
	spiltIndex := int(math.Floor(float64(len(records)) / 2))
	slice1 := records[0:spiltIndex]
	slice1Size, err := size(slice1)
	if err != nil {
		return [][]interface{}{}, err
	}
	if slice1Size < maxChunkSize {
		chunks = append(chunks, slice1)
	} else {
		slice1Chunks, err := spiltRecords(slice1)
		if err != nil {
			return [][]interface{}{}, err
		}
		chunks = append(chunks, slice1Chunks...)
	}
	slice2 := records[spiltIndex:]
	slice2Size, err := size(slice1)
	if err != nil {
		return [][]interface{}{}, err
	}
	if slice2Size < maxChunkSize {
		chunks = append(chunks, slice2)
	} else {
		slice2Chunks, err := spiltRecords(slice2)
		if err != nil {
			return [][]interface{}{}, err
		}
		chunks = append(chunks, slice2Chunks...)
	}

	return chunks, nil
}

func size(records []interface{}) (int, error) {
	r, err := json.Marshal(records)
	if err != nil {
		return 0, err
	}

	return len(r), nil
}

func (c *ClientProvider) send(channel string, records []interface{}) ([]*PutResponse, error) {
	inputs := make([]types.Record, 0)
	for _, r := range records {
		b, err := json.Marshal(r)
		if err != nil {
			return []*PutResponse{}, err
		}
		inputs = append(inputs, types.Record{Data: b})
	}

	params := &firehose.PutRecordBatchInput{
		DeliveryStreamName: aws.String(channel),
		Records:            inputs,
	}
	recordBatch, err := c.firehose.PutRecordBatch(context.Background(), params)
	if err != nil {
		return []*PutResponse{}, err
	}

	var res []*PutResponse
	for _, r := range recordBatch.RequestResponses {
		if r.RecordId != nil {
			response := &PutResponse{RecordID: *r.RecordId}
			if r.ErrorMessage != nil {
				response.Error = errors.New(*r.ErrorMessage)
			}
			res = append(res, response)
		}
	}
	return res, nil
}

type chanPutResponse struct {
	Result []*PutResponse
	Error  error
}

// DescribeDeliveryStream
// Describes the specified delivery stream and its status. For example, after your
// delivery stream is created, call DescribeDeliveryStream to see whether the
// delivery stream is ACTIVE and therefore ready for data to be sent to it. If the
// status of a delivery stream is CREATING_FAILED, this status doesn't change, and
// you can't invoke CreateDeliveryStream again on it. However, you can invoke the
// DeleteDeliveryStream operation to delete it. If the status is DELETING_FAILED,
// you can force deletion by invoking DeleteDeliveryStream again but with
// DeleteDeliveryStreamInput$AllowForceDelete set to true.
func (c *ClientProvider) DescribeDeliveryStream(channel string) (*DescribeOutput, error) {
	params := firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: &channel,
	}

	res, err := c.firehose.DescribeDeliveryStream(context.Background(), &params)
	if err != nil {
		return &DescribeOutput{}, err
	}
	return &DescribeOutput{StreamStatus: string(res.DeliveryStreamDescription.DeliveryStreamStatus)}, nil
}

// DescribeOutput ...
type DescribeOutput struct {
	StreamStatus string
}

// DeleteDeliveryStream ...
// Deletes a delivery stream and its data. To check the state of a delivery stream,
// use DescribeDeliveryStream. You can delete a delivery stream only if it is in
// one of the following states: ACTIVE, DELETING, CREATING_FAILED, or
// DELETING_FAILED. You can't delete a delivery stream that is in the CREATING
// state. While the deletion request is in process, the delivery stream is in the
// DELETING state. While the delivery stream is in the DELETING state, the service
// might continue to accept records, but it doesn't make any guarantees with
// respect to delivering the data. Therefore, as a best practice, first stop any
// applications that are sending records before you delete a delivery stream.
func (c *ClientProvider) DeleteDeliveryStream(channel string, force bool) error {
	params := firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: &channel,
	}
	if force {
		params.AllowForceDelete = &force
	}

	_, err := c.firehose.DeleteDeliveryStream(context.Background(), &params)
	if err != nil {
		return err
	}
	return nil
}
