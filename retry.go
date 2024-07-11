package hubur

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

const (
	// DefaultRetryTimes times of retry
	DefaultRetryTimes = 5
	// DefaultRetryDuration time duration of two retries
	DefaultRetryDuration = time.Second * 3
)

// RetryConfig is config for retry
type RetryConfig struct {
	context       context.Context
	retryTimes    int
	retryDuration time.Duration
}

// RetryFunc is function that retry executes
type RetryFunc func() error

// Option is for adding retry config
type Option func(*RetryConfig)

// RetryTimes set times of retry
func RetryTimes(n int) Option {
	return func(rc *RetryConfig) {
		rc.retryTimes = n
	}
}

// RetryDuration set duration of retries
func RetryDuration(d time.Duration) Option {
	return func(rc *RetryConfig) {
		rc.retryDuration = d
	}
}

// Context set retry context config
func Context(ctx context.Context) Option {
	return func(rc *RetryConfig) {
		rc.context = ctx
	}
}

// Retry executes the retryFunc repeatedly until it was successful or canceled by the context
// The default times of retries is 5 and the default duration between retries is 3 seconds
func Retry(retryFunc RetryFunc, opts ...Option) error {
	config := &RetryConfig{
		retryTimes:    DefaultRetryTimes,
		retryDuration: DefaultRetryDuration,
		context:       context.TODO(),
	}

	for _, opt := range opts {
		opt(config)
	}

	var i int
	for i < config.retryTimes || config.retryTimes < 0 {
		err := retryFunc()
		if err != nil {
			select {
			case <-time.After(config.retryDuration):
			case <-config.context.Done():
				return errors.New("retry is cancelled")
			}
		} else {
			return nil
		}
		i++
	}

	funcPath := runtime.FuncForPC(reflect.ValueOf(retryFunc).Pointer()).Name()
	lastSlash := strings.LastIndex(funcPath, "/")
	funcName := funcPath[lastSlash+1:]

	return fmt.Errorf("function %s run failed after %d times retry", funcName, i)
}
