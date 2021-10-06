package main

import (
	"flag"
	"github.com/LF-Engineering/insights-datasource-shared/firehose"
)

func init() {
	flag.StringVar(&region, "AWS_REGION", "", "The firehose region name")
}

func main() {
	jiraChannel := "jira"
	flag.Parse()

	// create new firehose client provider
	// you need to provide region as environment variable, or it will fall to default
	// which is us-east-1
	client, err := firehose.NewClientProvider()
	if err != nil {
		panic(err)
	}

	// use DescribeDeliveryStream to check the status of delivery stream
	describeOutput, err := client.DescribeDeliveryStream(jiraChannel)
	// check if deliver stream failed to be created or failed to be deleted, if so delete it
	if describeOutput.StreamStatus == "CREATING_FAILED" {
		err = client.DeleteDeliveryStream(jiraChannel, false)
		if err != nil {
			panic(err)
		}
	}

	if describeOutput.StreamStatus == "DELETING_FAILED" {
		err = client.DeleteDeliveryStream(jiraChannel, true)
		if err != nil {
			panic(err)
		}
	}

	if err != nil {
		// delivery stream is not exist, and we must create it
		// create new delivery stream channel named jira
		// you will need to create channel once, and then you can use it every time
		// to check if delivery stream is already exist you may use DescribeDeliveryStream
		err = client.CreateDeliveryStream(jiraChannel)
		if err != nil {
			panic(err)
		}
	}

	batches := generateBatch(10)

	// process batch records which will handle chunk size
	// it will check the batch record with firehose limitation
	// and separate the batch to accepted smaller chunks and process each one
	_, err = client.PutRecordBatch(jiraChannel, batches)
	if err != nil {
		panic(err)
	}

}

func generateBatch(count int) []interface{} {
	batch := make([]interface{}, 0)
	for i := 0; i < count; i++ {
		batch = append(batch, b)
	}
	return batch
}
