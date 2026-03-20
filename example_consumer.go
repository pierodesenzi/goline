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

    fmt.Printf("The content is %s\n", p)

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

	worker := provider.NewProvider(rdb, "queue5")
	ctx := context.Background()
	for {
		task, _ := worker.Next(ctx)
		function, ok := handlers[task.Function]
		if !ok { // If the function does not exist on handlers
			// TODO: send to DLQ
			continue
		}

		err := function(task.Params)
		if err != nil {
			// TODO: send to DLQ
			continue
		} else {
			worker.Ack(ctx, task.Id)
		}
	}
}
