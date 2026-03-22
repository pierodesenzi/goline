package provider

import (
	"github.com/redis/go-redis/v9"
	"fmt"
	"context"
	"encoding/json"
)

// Provider coordinates task consumption using Redis.
type Provider struct {
	rdb             *redis.Client
	queue           string
	processingQueue string
}

// Task represents a unit of work fetched from Redis.
type Task struct {
	Id       string
	Function string
	Params   json.RawMessage // Opaque JSON payload passed to the function
}

func NewProvider(rdb *redis.Client, queue string) *Provider {
	return &Provider{
		rdb:             rdb,
		queue:           "queue:" + queue,
		processingQueue: fmt.Sprintf("processing_queue:%s", queue),
	}
}

// Next blocks until a task is available, then:
// 1. Pops the task from the queue (BLPOP).
// 2. Deserializes the JSON payload into a Task.
// 3. Marks the task as "in-flight" by creating a processing key.
//
// This provides at-least-once delivery semantics: once popped,
// the task must be explicitly acknowledged via Ack, otherwise
// it remains "in-progress" (but not automatically retried).
func (p *Provider) Next(ctx context.Context) (*Task, []string, error) {
	raw, err := p.rdb.BLPop(ctx, 0, p.queue).Result()
	if err != nil {
		return nil, nil, err
	}

	var task Task
	if err := json.Unmarshal([]byte(raw[1]), &task); err != nil {
		return nil, raw, fmt.Errorf("invalid task payload: %w", err)
	}

	// Create a marker indicating this task is being processed.
	_, err = p.rdb.Set(ctx, p.processingQueue+":"+task.Id, 1, 0).Result()
	if err != nil {
		return nil, nil, err
	}

	return &task, nil, nil
}

// Ack acknowledges successful processing of a task by removing
// its in-flight marker. Uses UNLINK (non-blocking delete).
func (p *Provider) Ack(ctx context.Context, id string) error {
	return p.rdb.Unlink(ctx, p.processingQueue+":"+id).Err()
}


// DLQ:
type DLQItem struct {
	Task  *Task           `json:"task,omitempty"`  // present if parsing succeeded
	Raw   json.RawMessage `json:"raw,omitempty"`   // original payload if parsing failed
	Error string          `json:"error"`           // why it failed
}

// SendToDLQ stores a failed item (parsed or raw) in the DLQ.
func (p *Provider) SendToDLQ(ctx context.Context, key string, item DLQItem) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return p.rdb.RPush(ctx, key, data).Err()
}