package provider

import "encoding/json"

// Task represents a unit of work fetched from Redis.
type Task struct {
	Id       string
	Function string
	Params   json.RawMessage // Opaque JSON payload passed to the function
}

// DLQ:
type DLQItem struct {
	Task  *Task           `json:"task,omitempty"` // present if parsing succeeded
	Raw   json.RawMessage `json:"raw,omitempty"`  // original payload if parsing failed
	Error string          `json:"error"`          // why it failed
}
