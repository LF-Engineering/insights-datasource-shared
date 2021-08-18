package firehose

import (
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

func PutRecordBatch(c Config, channel string, data []byte) (int64,error) {

	sess := session.Must(session.NewSession())
	svr := firehose.New(sess, aws.NewConfig().WithRegion(c.Region).WithEndpoint(c.Endpoint))

	var records []*firehose.Record
	records = append(records, &firehose.Record{Data: data})
	params := &firehose.PutRecordBatchInput{
		DeliveryStreamName: aws.String(channel),
		Records:            records,
	}
	recordBatch, err := svr.PutRecordBatch(params)
	if err != nil {
		return 0, err
	}
	return *recordBatch.FailedPutCount, nil
}