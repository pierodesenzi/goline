package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupProvider(t *testing.T, queue string) (*Provider, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return NewProvider(rdb, queue), mr
}

func TestNext_Success(t *testing.T) {
	p, mr := setupProvider(t, "test")

	task := Task{
		Id:       "123",
		Function: "print",
		Params:   json.RawMessage(`{"a":1}`),
	}

	payload, _ := json.Marshal(task)
	mr.RPush("queue:test", string(payload))

	ctx := context.Background()
	res, raw, err := p.Next(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if raw != nil {
		t.Fatalf("expected raw to be nil on success")
	}

	if res.Id != "123" || res.Function != "print" {
		t.Fatalf("unexpected task: %+v", res)
	}

	// verify processing marker exists
	if !mr.Exists("processing_queue:test:123") {
		t.Fatal("expected processing marker to exist")
	}
}

func TestNext_InvalidJSON(t *testing.T) {
	p, mr := setupProvider(t, "test")

	mr.RPush("queue:test", "not-json")

	ctx := context.Background()
	res, raw, err := p.Next(ctx)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if res != nil {
		t.Fatal("expected nil task on error")
	}

	if raw == nil {
		t.Fatal("expected raw payload to be returned")
	}
}

func TestAck_RemovesProcessingMarker(t *testing.T) {
	p, mr := setupProvider(t, "test")

	mr.Set("processing_queue:test:123", "1")

	ctx := context.Background()
	err := p.Ack(ctx, "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mr.Exists("processing_queue:test:123") {
		t.Fatal("expected processing marker to be removed")
	}
}

func TestSendToDLQ_Success(t *testing.T) {
	p, mr := setupProvider(t, "test")

	item := DLQItem{
		Error: "something failed",
	}

	ctx := context.Background()
	err := p.SendToDLQ(ctx, "dlq:test", item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values, err := mr.List("dlq:test")
	if len(values) != 1 {
		t.Fatalf("expected 1 item in DLQ, got %d", len(values))
	}

	var stored DLQItem
	if err = json.Unmarshal([]byte(values[0]), &stored); err != nil {
		t.Fatalf("failed to unmarshal DLQ item: %v", err)
	}

	if stored.Error != "something failed" {
		t.Fatalf("unexpected DLQ item: %+v", stored)
	}
}

func TestCheckQueue(t *testing.T) {
	p, mr := setupProvider(t, "test")

	mr.RPush("queue:test", "a", "b", "c")

	ctx := context.Background()
	tasks, err := p.CheckQueue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
}