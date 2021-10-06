# firehose

#### Firehose has the following functionality:

1. `CreateDeliveryStream`
2. `PutRecord`
3. `PutRecordBatch`
4. `DescribeDeliveryStream`
5. `DeleteDeliveryStream`

### Use cases

1- insert single record to existing delivery stream:
    - first you will need to create new firehose client
    - use `PutRecord` to insert a single record to specific stream, which take 2 params `channel` and `record`

2- insert bulk records to new delivery stream:
    - first you will need to create new firehose client
    - then create new delivery stream using `CreateDeliveryStream`, which take 1 param `channelName`
    - use `PutRecordBatch` to insert multiple records to specific stream, which take 2 params `channel` and `bulkrecords`
        PutRecordBatch will handle size limitations internally, by dividing batch to smaller batches with proper size.

3- check existence of delivery stream:
    - use `DescribeDeliveryStream` 
        if status is ACTIVE therefore ready for data to be sent to it
        if status is CREATING_FAILED, so you need to delete deliver stream and create it again
