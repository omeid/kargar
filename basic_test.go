package kargar_test

import (
	"sync"
	"testing"
	"time"

	"github.com/omeid/gonzo/context"
	"github.com/omeid/kargar"
)

func TestCancel(t *testing.T) {

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
		b.Context().Error(err)
	}
}

func TestConcurrency(t *testing.T) {

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
}
