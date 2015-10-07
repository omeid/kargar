# Kargar [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/omeid/kargar) 
<p align="center">
<img width="100%" src="https://talks.golang.org/2012/waza/gophercomplex6.jpg">
</p>

Kagrar is a concurrency-aware task harness and dependency management with first-class support for Deadlines, Cancelation, and task-labeled logging.


Kargar allows you to cancel tasks.

```go
ctx, cancel := context.WithCancel(context.Background())

b := kargar.NewBuild(ctx)

b.Add(kargar.Task{

	Name:  "say-hello",
	Usage: "This tasks is self-documented, it says hello for every second.",

	Action: func(ctx context.Context) error {

		second := time.NewTicker(time.Second)

		for {
			select {

			case <-second.C:
				ctx.Info("Hello!")
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	},
})

go func() {
	time.Sleep(4 * time.Second)
	cancel()
}()

err := b.Run("say-hello")
if err != nil {
	log.Println(err)
}
```

```sh
INFO[0001] Hello!                                        task=say-hello
INFO[0002] Hello!                                        task=say-hello
INFO[0003] Hello!                                        task=say-hello
INFO[0004] Hello!                                        task=say-hello
ERRO[0004] context canceled
```

To avoid race-conditions, there will be one and only one instance of any given task running at any given time.


```go
b := kargar.New()

b.Add(
	kargar.Task{

		Name:  "slow-dependency",
		Usage: "This is a slow task, takes 3 seconds before it prints time.",

		Action: func(ctx context.Context) error {

			ctx.Warn("I am pretty slow.")
			for {
				select {

				case now := <-time.After(3 * time.Second):
					ctx.Infof("Time is %s", now.Format(time.Kitchen))
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		},
	},

	kargar.Task{
		Name:   "a",
		Usage:  "This task depends on 'slow-dependency.'",
		Deps:   []string{"slow-dependency"},
		Action: kargar.Noop(),
	},

	kargar.Task{
		Name:   "b",
		Usage:  "This task depends on 'slow-dependency.'",
		Deps:   []string{"slow-dependency"},
		Action: kargar.Noop(),
	},

	kargar.Task{
		Name:   "c",
		Usage:  "This task depends on 'slow-dependency.'",
		Deps:   []string{"slow-dependency"},
		Action: kargar.Noop(),
	},
)

ctx := b.Context()
var wg sync.WaitGroup

for _, t := range []string{"a", "b", "c"} {
	wg.Add(1)
	go func(t string) {
		defer wg.Done()
		err := b.Run(t)
		if err != nil {
			ctx.Error(err)
		}
	}(t)
}

wg.Wait()
```

```
WARN[0000] I am pretty slow.                             parent=c task=slow-dependency
INFO[0003] Time is 2:14PM                                parent=c task=slow-dependency
WARN[0003] I am pretty slow.                             parent=b task=slow-dependency
INFO[0006] Time is 2:14PM                                parent=b task=slow-dependency
WARN[0006] I am pretty slow.                             parent=a task=slow-dependency
INFO[0009] Time is 2:14PM                                parent=a task=slow-dependency
```

# Kar

To build and run Kargar tasks from CLI, see [kar](https://github.com/omeid/kar).

### LICENSE
  [MIT](LICENSE).
