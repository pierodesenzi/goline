package main

import (
    "github.com/pierodesenzi/goline/provider"
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

	queue := "queue5"

	worker := provider.NewProvider(rdb, queue)
	ctx := context.Background()
	for {
		task, raw, err := worker.Next(ctx)
		if err != nil {  // invalid payload
			worker.SendToDLQ(ctx, queue + ":dlq:invalid", provider.DLQItem{
				Raw: json.RawMessage(raw[1]),
				Error: "invalid payload",
			})
			continue
		}
		function, ok := handlers[task.Function]
		if !ok { // If the function does not exist on handlers
			worker.SendToDLQ(ctx, queue + ":dlq:function_not_found", provider.DLQItem{
				Task: task,
				Error: "function not present",
			})
			continue
		}

		err = function(task.Params)
		if err != nil {
			worker.SendToDLQ(ctx, queue + ":dlq:error_executing", provider.DLQItem{
				Task: task,
				Error: "error during function execution",
			})
		} else {
			worker.Ack(ctx, task.Id)
		}
	}
}
