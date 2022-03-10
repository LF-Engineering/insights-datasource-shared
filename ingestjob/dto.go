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

// TaskTopHits result
type TaskTopHits struct {
	Hits TaskHits `json:"hits"`
}

// TaskHits result
type TaskHits struct {
	Hits []TaskNestedHits `json:"hits"`
}

// TaskNestedHits is the actual hit data
type TaskNestedHits struct {
	ID     string  `json:"_id"`
	Source TaskLog `json:"_source"`
}

type TaskLog struct {
	Id            string              `json:"id"`            // task id
	Connector     string              `json:"connector"`     // name of connector as https://github.com/LF-Engineering/lfx-event-schema/blob/main/service/insights/shared.go#L32-L49
	EndpointId    string              `json:"endpoint_id"`   // generated ID of the endpoint per lfx-event-schema generator functions
	Configuration []map[string]string `json:"configuration"` // metadata
	CreatedAt     *time.Time          `json:"created_at,omitempty"`
}
