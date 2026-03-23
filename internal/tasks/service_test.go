package tasks

import (
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTest(t *testing.T) (*Service, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return NewService(rdb), mr
}

func TestCreateQueue(t *testing.T) {
	svc, _ := setupTest(t)

	resp, err := svc.Create("test-queue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != "CREATED" {
		t.Fatalf("expected CREATED, got %s", resp.Status)
	}

	// second call should be idempotent
	resp, err = svc.Create("test-queue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != "ALREADY_EXISTS" {
		t.Fatalf("expected ALREADY_EXISTS, got %s", resp.Status)
	}
}

func TestEnqueue_NonExistingQueue(t *testing.T) {
	svc, _ := setupTest(t)

	resp, err := svc.Enqueue("missing-queue", "fn", map[string]any{"a": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != "QUEUE_DOES_NOT_EXIST" {
		t.Fatalf("expected QUEUE_DOES_NOT_EXIST, got %s", resp.Status)
	}
}

func TestEnqueue_Success(t *testing.T) {
	svc, mr := setupTest(t)

	_, err := svc.Create("test-queue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := svc.Enqueue("test-queue", "print", map[string]any{"b": 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != "ENQUEUED" {
		t.Fatalf("expected ENQUEUED, got %s", resp.Status)
	}

	if resp.Id == "" {
		t.Fatal("expected non-empty task ID")
	}

	// verify Redis state
	values, err := mr.List("queue:test-queue")
	if len(values) != 1 {
		t.Fatalf("expected 1 task in queue, got %d", len(values))
	}

	var task map[string]any
	if err = json.Unmarshal([]byte(values[0]), &task); err != nil {
		t.Fatalf("failed to unmarshal task: %v", err)
	}

	if task["function"] != "print" {
		t.Fatalf("expected function=print, got %v", task["function"])
	}
}

func TestCheckQueue_NonExisting(t *testing.T) {
	svc, _ := setupTest(t)

	resp, err := svc.CheckQueue("missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != "QUEUE_DOES_NOT_EXIST" {
		t.Fatalf("expected QUEUE_DOES_NOT_EXIST, got %s", resp.Status)
	}

	if len(resp.Tasks) != 0 {
		t.Fatalf("expected empty tasks, got %v", resp.Tasks)
	}
}

func TestCheckQueue_WithTasks(t *testing.T) {
	svc, _ := setupTest(t)

	_, err := svc.Create("test-queue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Enqueue("test-queue", "fn1", map[string]any{"x": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Enqueue("test-queue", "fn2", map[string]any{"y": 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := svc.CheckQueue("test-queue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != "OK" {
		t.Fatalf("expected OK, got %s", resp.Status)
	}

	if len(resp.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(resp.Tasks))
	}
}