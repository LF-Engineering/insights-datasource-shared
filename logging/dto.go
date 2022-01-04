package logging

import "time"

// Log ...
type Log struct {
	Datasource string    `json:"datasource"`
	Endpoint   string    `json:"endpoint"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// TopHits result
type TopHits struct {
	Hits Hits `json:"hits"`
}

// Hits result
type Hits struct {
	Hits []NestedHits `json:"hits"`
}

// NestedHits is the actual hit data
type NestedHits struct {
	ID     string `json:"_id"`
	Source Log    `json:"_source"`
}
