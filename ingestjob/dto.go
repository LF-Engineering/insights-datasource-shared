package ingestjob

import "time"

// Log ...
type Log struct {
	Connector     string              `json:"connector"`
	Configuration []map[string]string `json:"configuration"`
	Status        string              `json:"status"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	Message       string              `json:"message"`
	From          *time.Time          `json:"from,omitempty"`
	To            *time.Time          `json:"to,omitempty"`
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
