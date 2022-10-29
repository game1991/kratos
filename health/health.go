package health

import (
	"context"
	"sync"
	"time"
)

type Status uint8

const (
	Down = iota
	Up
)

type Checker interface {
	Check(ctx context.Context) (map[string]interface{}, error)
}

type Health struct {
	status   Status
	checkers map[string]Checker
	rwLock   sync.RWMutex
	timeout  time.Duration
}

func (h *Health) HealthCheckAll(ctx context.Context) (Result, error) {
	r := Result{
		Status:     h.status,
		Components: make(map[string]ComponentResult),
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	h.rwLock.RLock()
	defer h.rwLock.RUnlock()
	for component, checker := range h.checkers {
		wg.Add(1)
		go func(ctx context.Context, component string, checker Checker) {
			defer wg.Done()
			res := h.check(ctx, checker)
			mu.Lock()
			r.Components[component] = res
			mu.Unlock()
		}(ctx, component, checker)
	}
	wg.Wait()
	return r, nil
}

func (h *Health) HealthCheckOne(ctx context.Context, component string) (ComponentResult, error) {
	h.rwLock.RLock()
	checker, ok := h.checkers[component]
	h.rwLock.RUnlock()
	if !ok {
		return ComponentResult{}, ErrComponentNotFound
	}

	return h.check(ctx, checker), nil
}

func (h *Health) check(ctx context.Context, checker Checker) ComponentResult {
	resCh := make(chan ComponentResult, 1)
	defer close(resCh)
	go func() {
		c := ComponentResult{}
		details, err := checker.Check(ctx)
		if err != nil {
			c = ComponentResult{
				Status:  Down,
				Err:     err,
				Details: details,
			}
		} else {
			c = ComponentResult{
				Status:  Up,
				Err:     nil,
				Details: details,
			}
		}
		resCh <- c
	}()
	select {
	case <-time.After(h.timeout):
		return ComponentResult{
			Status: Down,
			Err:    context.DeadlineExceeded,
		}
	case res := <-resCh:
		return res
	}
}
