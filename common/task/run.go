package task

import (
	"context"
	"fmt"
)

func OnSuccess(f func() error, g func() error) func() error {
	return func() error {
		if err := f(); err != nil {
			return err
		}
		return g()
	}
}

func Run(ctx context.Context, tasks ...func() error) error {
	n := len(tasks)
	done := make(chan error, 1)
	channel := make(chan struct{}, n)

	for _, task := range tasks {
		go func(f func() error) {
			err := f() // run requestDonePost(which runs requestDone) & responseDone
			if err == nil {
				channel <- struct{}{}
				return
			}

			select {
			case done <- err:
			default:
			}
		}(task)
	}

	for i := 0; i < n; i++ {
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return fmt.Errorf("%w:%w", ctx.Err(), context.Cause(ctx))
		case <-channel:
		}
	}

	return nil
}
