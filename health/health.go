package health

import (
	"context"
	"sync"
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
}

func (h *Health) HealthCheckAll(ctx context.Context) (Result, error) {
	r := Result{
		Status:  h.status,
		Details: make([]ComponentResult, 0, len(h.checkers)),
	}
	lock := sync.Mutex{}

	var wg sync.WaitGroup
	wg.Add(len(h.checkers))

	h.rwLock.RLock()
	defer h.rwLock.RUnlock()
	for component, checker := range h.checkers {
		go func(component string, checker Checker) {
			defer wg.Done()
			cr := h.check(ctx, component, checker)
			lock.Lock()
			r.Details = append(r.Details, cr)
			lock.Unlock()
		}(component, checker)
	}
	wg.Wait()
	return r, nil
}

func (h *Health) HealthCheckOne(ctx context.Context, component string) (ComponentResult, error) {
	h.rwLock.RLock()
	checker, ok := h.checkers[component]
	h.rwLock.RUnlock()
	if !ok || checker == nil {
		return ComponentResult{}, ErrComponentNotFound
	}

	return h.check(ctx, component, checker), nil
}

func (h *Health) check(ctx context.Context, component string, c Checker) ComponentResult {
	details, err := c.Check(ctx)
	if err != nil {
		return ComponentResult{
			Name:     component,
			Status:   Down,
			ErrorMsg: err.Error(),
			Details:  details,
		}
	}
	return ComponentResult{
		Name:     component,
		Status:   Up,
		ErrorMsg: "",
		Details:  details,
	}
}
