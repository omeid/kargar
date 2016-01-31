package kargar

import (
	"sync"

	"github.com/omeid/gonzo/context"
)

// Action is a function that is called when a task is run.
type Action func(context.Context) error

// Noop is an Action that does nothing and returns `nil` immediately.
// For clarity and to avoid weird bugs, every task must have an Action
// Noop is provided for tasks that are only used to group a collection
// of tasks as dependency.
func Noop() Action {
	return func(ctx context.Context) error { return nil }
}

// Task holds the meta information and an action.
type Task struct {
	// Taks name
	Name string
	// A short description of the task.
	Usage string
	// A long explanation of how the task works.
	Description string
	// List of dependencies.
	// When running a task, the dependencies will be run in the order
	Deps []string
	// The function to call when the task is invoked.
	Action Action
}

type task struct {
	Task
	deps    taskstack
	lock    sync.Mutex
	done    <-chan struct{}
	running bool
}

type taskstack map[string]*task

func (t *task) run(ctx context.Context) error {
	t.lock.Lock()
	defer func() {
		t.running = false
		t.lock.Unlock()
	}()

	if task, ok := ctx.Value("task").(string); ok {
		ctx = context.WithValue(ctx, "parent", task)
	}

	if t.Name != "default" {
		ctx = context.WithValue(ctx, "task", t.Name)
	}

	ctx, cancel := context.WithCancel(ctx)

	var once sync.Once;

	ctx.Debug("start")
	var wg sync.WaitGroup
	for _, t := range t.deps {
		select {
		case <-ctx.Done():
			break
		default:
			wg.Add(1)
			go func(t *task) {
				defer wg.Done()
				ctx.Debug("Waiting for %s", t.Name)
				err := t.run(ctx)
				if err != nil {
					once.Do(func() {
						ctx.Warnf("%s failed. Giving up!", t.Name);
					});

					cancel()
					if err != context.Canceled {
						ctx.Error(err)
					}
				}
			}(t)
		}
	}
	wg.Wait()

	err := ctx.Err()
	if err != nil {
		if err == context.Canceled {
			return nil
		}
		return err
	}

	t.running = true
	err = t.Action(ctx)
	if err == nil {
		ctx.Debug("Done.")
	}
	return err
}
