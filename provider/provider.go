package provider

import (
	"github.com/redis/go-redis/v9"
	"fmt"
	"context"
	"encoding/json"
)

type Provider struct {
	rdb             *redis.Client
	queue           string
	processingQueue string
}


type Task struct {
	queue	string
	id		string
	params	map[string]any
}

func NewProvider(rdb *redis.Client, queue string) *Provider {
	return &Provider{
		rdb:             rdb,
		queue:           queue,
		processingQueue: fmt.Sprintf("%s:processing", queue),
	}
}

func (p *Provider) Next(ctx context.Context) (*Task, error) {
	raw, err := p.rdb.BLMove(ctx, p.queue, p.processingQueue, "RIGHT", "LEFT", 0).Result()
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal([]byte(raw), &task); err != nil {
		return nil, fmt.Errorf("invalid task payload: %w", err)
	}

	return &task, nil
}

/*
TODO: implement Ack
func (q *RedisQueue) Ack(ctx context.Context, raw string) error {
	return q.rdb.LRem(ctx, q.processingKey, 1, raw).Err()
}
*/