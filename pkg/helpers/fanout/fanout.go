package fanout

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
)

// Item is fan out general item
type Item struct {
	result interface{}
	err    error
}

// ConcurrentResponseItems process with items channel
func ConcurrentResponseItems(ctx context.Context, worker func(ctx context.Context, value interface{}) (interface{}, error), parameters ...interface{}) chan Item {

	var wg sync.WaitGroup

	out := make(chan Item, len(parameters))

	for _, param := range parameters {
		wg.Add(1)

		go func(param interface{}) {
			defer wg.Done()
			result, err := worker(ctx, param)
			out <- Item{
				result: result,
				err:    err,
			}
		}(param)

	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out

}

// ConcurrentResponseErrors process with errors channel
func ConcurrentResponseErrors(ctx context.Context, worker func(ctx context.Context, value interface{}) error, parameters ...interface{}) chan error {

	var wg sync.WaitGroup

	out := make(chan error, len(parameters))

	for _, param := range parameters {
		wg.Add(1)

		go func(param interface{}, wg *sync.WaitGroup, out chan error) {
			defer wg.Done()
			err := worker(ctx, param)
			if err != nil {
				out <- err
			}
		}(param, &wg, out)

	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out

}

// ReadItemChannel reads items channel and constructs error
func ReadItemChannel(ch chan Item) ([]interface{}, error) {
	multiErr := &multierror.Error{}

	var results []interface{}

	for result := range ch {
		if result.err != nil {
			multiErr = multierror.Append(multiErr, result.err)
		} else {
			results = append(results, result.result)
		}

	}

	return results, multiErr.ErrorOrNil()

}

// ReadErrorsChannel reads error's channel and constructs error
func ReadErrorsChannel(ch chan error) error {
	multiErr := &multierror.Error{}

	for err := range ch {
		multiErr = multierror.Append(multiErr, err)
	}
	return multiErr.ErrorOrNil()

}
