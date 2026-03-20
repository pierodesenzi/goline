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

	key := "active_queues:" + queue

	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		// Initialize empty list with only a head entry
		if err := s.rdb.Set(ctx, key, 1, 0).Err(); err != nil {
			return nil, err
		}
	} else {
		// Cannot recreate existing queue
		return map[string]interface{}{
			"queue":  queue,
			"status": "ALREADY_EXISTS",
		}, nil
	}

	return map[string]interface{}{
		"queue":  queue,
		"status": "CREATED",
	}, nil
}

func (s *Service) Enqueue(queue string, function string, params map[string]any) (map[string]interface{}, error) {
	// Pushes a task into the queue

	// Fails if operation takes more than 2 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// enqueuing a task should not create a queue in Redis  
	exists, err := s.rdb.Exists(ctx, "active_queues:" + queue).Result()
	if err != nil {
		return nil, err
	}

	if exists == 0 {
		return map[string]interface{}{
			"queue":  "",
			"id": "",
			"status": "QUEUE_DOES_NOT_EXIST",
		}, nil
	}

	// generating task JSON
	key := "queue:" + queue
	task_uuid := uuid.NewString()
	task := map[string]any{
		"id": task_uuid,
		"function": function,
		"params": params,
	}

	// convert task to string format
	payload, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	// FIFO queue: push to the right
	if err := s.rdb.RPush(ctx, key, payload).Err(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"queue":  queue,
		"id": task_uuid,
		"status": "ENQUEUED",
	}, nil
}