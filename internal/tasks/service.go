package tasks

import (
	"context"
	"encoding/json"
	"time"
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
)

type Service struct {
	rdb *redis.Client
}

func NewService(rdb *redis.Client) *Service {
	return &Service{
		rdb: rdb,  // injecting the Redis client
	}
}

func (s *Service) Create(queue string) (map[string]interface{}, error) {
	// Initializes an empty queue

	// Fails if operation takes more than 2 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := "queue:" + queue

	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		// Initialize empty list with only a head entry
		if err := s.rdb.RPush(ctx, key, "__init__").Err(); err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"queue":  queue,
		"status": "created",
	}, nil
}

func (s *Service) Enqueue(queue string, params map[string]interface{}) (map[string]interface{}, error) {
	// Pushes a task into the queue

	// Fails if operation takes more than 2 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := "queue:" + queue

	task_uuid := uuid.NewString()
	params["id"] = task_uuid

	payload, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}


	// enqueuing a task should not create a queue in Redis  
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if exists == 0 {
		return map[string]interface{}{
			"queue":  "",
			"id": "",
			"status": "not enqueued - queue not found",
		}, nil
	}

	// FIFO queue: push to the right
	if err := s.rdb.RPush(ctx, key, payload).Err(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"queue":  queue,
		"id": task_uuid,
		"status": "enqueued",
	}, nil
}