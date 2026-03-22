package tasks

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	rdb *redis.Client
}

func NewService(rdb *redis.Client) *Service {
	return &Service{
		rdb: rdb, // injecting the Redis client
	}
}

type CreateResponse struct {
	Queue  string `json:"queue"`
	Status string `json:"status"`
}

type EnqueueResponse struct {
	Id     string `json:"id"`
	Queue  string `json:"queue"`
	Status string `json:"status"`
}

// Create signals the creation of an empty queue
func (s *Service) Create(queue string) (CreateResponse, error) {
	log.Printf("Create called for queue=%s", queue)

	// Fails if operation takes more than 2 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := "active_queues:" + queue

	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("Create failed checking existence for queue=%s: %v", queue, err)
		return CreateResponse{}, err
	}

	if exists == 0 { // queue does not exist, so create its existence signal
		if err := s.rdb.Set(ctx, key, 1, 0).Err(); err != nil {
			log.Printf("Create failed setting key for queue=%s: %v", queue, err)
			return CreateResponse{}, err
		}
		log.Printf("Queue created: %s", queue)
	} else { // queue does exist, don't recreate it
		log.Printf("Create skipped, queue already exists: %s", queue)
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

// Enqueue pushes a task into the queue
func (s *Service) Enqueue(queue string, function string, params map[string]any) (EnqueueResponse, error) {
	log.Printf("Enqueue called for queue=%s function=%s", queue, function)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	exists, err := s.rdb.Exists(ctx, "active_queues:"+queue).Result()
	if err != nil {
		log.Printf("Enqueue failed checking queue existence for queue=%s: %v", queue, err)
		return EnqueueResponse{}, err
	}

	if exists == 0 { // creating a task for a non existing queue should not create it
		log.Printf("Enqueue rejected, queue does not exist: %s", queue)
		return EnqueueResponse{
			Id:     "",
			Queue:  "",
			Status: "QUEUE_DOES_NOT_EXIST",
		}, nil
	}

	key := "queue:" + queue
	taskUUID := uuid.NewString()

	// creating task map to store in Redis
	task := map[string]any{
		"id":       taskUUID,
		"function": function,
		"params":   params,
	}

	payload, err := json.Marshal(task)
	if err != nil {
		log.Printf("Enqueue failed marshaling task for queue=%s: %v", queue, err)
		return EnqueueResponse{}, err
	}

	// pushing to the queue
	if err := s.rdb.RPush(ctx, key, payload).Err(); err != nil {
		log.Printf("Enqueue failed pushing to queue=%s: %v", queue, err)
		return EnqueueResponse{}, err
	}

	log.Printf("Task enqueued: queue=%s id=%s", queue, taskUUID)

	return EnqueueResponse{
		Id:     taskUUID,
		Queue:  queue,
		Status: "ENQUEUED",
	}, nil
}
