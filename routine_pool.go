package goroutine_pool

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
)

type content struct {
	work func() error
	end  *struct{}
}

func work(w func() error) content {
	return content{work: w}
}

func end() content {
	return content{end: &struct{}{}}
}

type RoutinePool struct {
	capacity uint
	ch       chan content
}

func NewRoutinePool(ctx context.Context, capacity uint) *RoutinePool {
	ch := make(chan content)
	pool := RoutinePool{
		capacity: capacity,
		ch:       ch,
	}

	for i := uint(0); i < capacity; i++ {
		safeGo(ctx, func() {
			for {
				select {
				case cont := <-ch:
					if cont.end != nil {
						return
					}

					if cont.work != nil {
						if err := cont.work(); err != nil {
							log.Printf("run work failed: %v", err)
						}
					}
				}
			}
		})
	}

	return &pool
}

func (pool *RoutinePool) Submit(w func() error) {
	pool.ch <- work(w)
}

func (pool *RoutinePool) Shutdown() {
	defer close(pool.ch)
	for i := uint(0); i < pool.capacity; i++ {
		pool.ch <- end()
	}
}

func safeGo(ctx context.Context, f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				content := fmt.Sprintf("Safe Go Capture Panic In Go Groutine\n%s", string(debug.Stack()))
				log.Fatal(ctx, content)
			}
		}()

		f()
	}()
}
