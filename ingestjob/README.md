# ingest job
ingestjob library used by connectors to handle logging

### How it works :
* when connector start it should create new log record using `Write` function with a unique identity based on
`connector`, `configuration` and `creation date` with status in progress.

* once connector finished it should call `Write` function again with the same identity `connector`, `configuration`
and `creation date` which internally update the same log record to done status.

* in case connector failed to complete its run, it should call `Write` function again with the same identity which
internally update the same log record to failed status.


### Ingest job has 4 functionalities:

1- `Write` create a log record if it is not exist. if record exists update it.

2- `Read` get a list of logs based on connector and status.

3- `Count` which count connector logs in a specific status.

4- `Filter` which filter log records based on any of 
`status`, `configuration`, `from` and `to` date.
