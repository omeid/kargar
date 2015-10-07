package kargar

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/omeid/gonzo/context"
)

var ctx = context.Background()

//Meta holds information about the build.
type Meta struct {
	// The name of the program. Defaults to os.Args[0]
	Name string
	// Description of the program.
	Usage string
	// Version of the program
	Version string
	// Author
	Author string
	// Author e-mail
	Email string

	// License
	License string
}

// Build is a simple build harness that you can register tasks and their
// dependencies and then run them.
type Build struct {
	Meta Meta
	ctx  context.Context

	tasks taskstack

	//cleanups    []func()
	//runcleanups bool

	lock sync.Mutex
}

// New returns a Build with a contex with no deadline or values and is never canceled.
func New() *Build {
	return NewBuild(
		context.Background(),
	)
}

// NewBuild returns a Build using the provided Context.
func NewBuild(ctx context.Context) *Build {
	//done := make(chan struct{})
	return &Build{
		ctx:   ctx,
		tasks: make(taskstack),
		lock:  sync.Mutex{},
	}
}

func (b *Build) Context() context.Context {
	return b.ctx
}

// Add registers the provided tasks to the build.
// Circular Dependencies are not allowed.
func (b *Build) Add(tasks ...Task) error {

	b.lock.Lock()
	defer b.lock.Unlock()
	for i, T := range tasks {
		if T.Name == "" {
			return fmt.Errorf("Task %d Missing Name.", i)
		}

		if T.Action == nil {
			return fmt.Errorf("Task %s Missing Action.", T.Name)
		}

		if T.Usage == "" {
			return fmt.Errorf("Task %s Missing Usage.", T.Name)
		}

		if _, ok := b.tasks[T.Name]; ok {
			return fmt.Errorf("Duplicate task: %s", T.Name)
		}
		t := &task{Task: T, deps: make(taskstack), running: false}

		for _, dep := range t.Deps {
			d, ok := b.tasks[dep]
			if !ok {
				return fmt.Errorf("Missing Task %s. Required by Task %s.", dep, t.Name)
			}
			_, ok = d.deps[t.Name]
			if ok {
				return fmt.Errorf("Circular dependency %s requies %s and around.", d.Name, t.Name)
			}
			t.deps[dep] = d
		}

		b.tasks[t.Name] = t
	}
	return nil
}

var ErrorNoSuchTask = fmt.Errorf("No Such Task.")

//RunFor runs a task using an alternative context.
//This this is typically useful when you want to dynamically
//invoked tasks from another task but still maintain proper
//context hireachy.
func (b *Build) RunFor(ctx context.Context, tasks ...string) error {

	if !kargar() {
		return errors.New("KARGAR=false, escaping run")
	}

	for _, name := range tasks {
		select {
		case <-ctx.Done():
			b.ctx.Warn("Build Canacled.")

		default:
			t, ok := b.tasks[name]
			if !ok {
				return ErrorNoSuchTask
			}
			err := t.run(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//Run runs the provided lists of tasks.
func (b *Build) Run(tasks ...string) error {
	return b.RunFor(b.ctx, tasks...)
}

func kargar() bool {
	kargar := os.Getenv("KARGAR")
	if kargar == "" {
		return true
	}
	k, err := strconv.ParseBool(kargar)
	if err != nil {
		ctx.Fatal(err)
	}
	return k
}
