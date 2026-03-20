package main

import (
    "pierodesenzi/goline/provider"
	"github.com/redis/go-redis/v9"
	"encoding/json"

	"fmt"

	"context"
)

// define functions
type printContentParams struct {
	Function string `json:"function"`
}

func printContent(raw json.RawMessage) error {
	var p printContentParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}

    fmt.Printf("The content is %s", p)

	return nil
}

type HandlerFunc func(json.RawMessage) error

var handlers = map[string]HandlerFunc{
	"print_content": printContent,
}

func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	worker := provider.NewProvider(rdb, "queue1")
	ctx := context.Background()
	for {
		task, _ := worker.Next(ctx)
		handlers[task.Function](task.Params)
	}

	// TODO: implement Ack() for success case
	// q.Ack(ctx, raw)

}
