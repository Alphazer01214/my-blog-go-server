package keepalive

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
}

type Manager struct {
	tasks  []*Task
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (m *Manager) Register(name string, interval time.Duration, fn func(ctx context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = append(m.tasks, &Task{
		Name:     name,
		Interval: interval,
		Fn:       fn,
	})
}

func (m *Manager) Start() {
	m.mu.Lock()
	tasks := m.tasks
	m.mu.Unlock()
	for _, t := range tasks {
		m.wg.Add(1)
		go m.runTask(t)
	}
	fmt.Println("[keepalive] all tasks started")
}

func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
	fmt.Println("[keepalive] all tasks stopped")
}

func (m *Manager) runTask(t *Task) {
	defer m.wg.Done()

	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	if err := t.Fn(m.ctx); err != nil {
		fmt.Printf("[keepalive] %s initial run error: %v\n", t.Name, err)
	}

	for {
		select {
		case <-m.ctx.Done():
			fmt.Printf("[keepalive] %s stopped\n", t.Name)
			return
		case <-ticker.C:
			if err := t.Fn(m.ctx); err != nil {
				fmt.Printf("[keepalive] %s run error: %v\n", t.Name, err)
			}
		}
	}
}
