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

type CreateResponse struct {
	Queue	string	`json:"queue"`
	Status	string	`json:"status"`
}

type EnqueueResponse struct {
	Id		string	`json:"id"`
	Queue	string	`json:"queue"`
	Status	string	`json:"status"`
}

func (s *Service) Create(queue string) (CreateResponse, error) {
	// Initializes an empty queue

	// Fails if operation takes more than 2 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := "active_queues:" + queue

	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return CreateResponse{}, err
	}
	if exists == 0 {
		// Initialize empty list with only a head entry
		if err := s.rdb.Set(ctx, key, 1, 0).Err(); err != nil {
			return CreateResponse{}, err
		}
	} else {
		// Cannot recreate existing queue
		return CreateResponse{
			Queue:  queue,
			Status: "ALREADY_EXISTS",
		}, nil
	}

	return CreateResponse{
		Queue:  queue,
		Status: "CREATED",
	}, nil
}

func (s *Service) Enqueue(queue string, function string, params map[string]any) (EnqueueResponse, error) {
	// Pushes a task into the queue

	// Fails if operation takes more than 2 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// enqueuing a task should not create a queue in Redis  
	exists, err := s.rdb.Exists(ctx, "active_queues:" + queue).Result()
	if err != nil {
		return EnqueueResponse{}, err
	}

	if exists == 0 {
		return EnqueueResponse{
			Id:  "",
			Queue: "",
			Status: "QUEUE_DOES_NOT_EXIST",
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

	// convert task to bytes format
	payload, err := json.Marshal(task)
	if err != nil {
		return EnqueueResponse{}, err
	}

	// FIFO queue: push to the right
	if err := s.rdb.RPush(ctx, key, payload).Err(); err != nil {
		return EnqueueResponse{}, err
	}

	return EnqueueResponse{
		Id: task_uuid,
		Queue:  queue,
		Status: "ENQUEUED",
	}, nil
}