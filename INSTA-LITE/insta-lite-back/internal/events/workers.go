package events

import (
	"context"
	"sync"
)

type Handler[T any] func(ctx context.Context, e T)

func StartWorkers[T any](ctx context.Context, wg *sync.WaitGroup, bus *TypedBus[T], numWorkers int, handle Handler[T]) {
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case e := <-bus.Events():
					handle(ctx, e)
				}
			}
		}()
	}
}
