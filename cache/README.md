# caching
caching library used by connectors

### How it works :
connector should check if `IsKeyCreated` if key is not already created, connector should `CreateCache`.

Caching has 2 functionalities:

1- `IsKeyCreated` it takes `objectID` which check if the provided key is already cached, key is a full path
 for connectors it is `cache/{connectorName}/{objectID}`

2- `CreateCache` which store a cache record, it takes `[]map[string]interface` each map
contain 2 keys `id` and `data`.id is the `objectID`. data is the bytes of the actual object.

3- `GetLastSync` which get connector last sync date

4- `SetLastSync` which update connector last sync date
