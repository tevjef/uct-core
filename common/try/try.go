// https://github.com/matryer/try
package try

import (
	"math"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

// MaxRetries is the maximum number of retries before bailing.
var MaxRetries = 10

var errMaxRetriesReached = errors.New("exceeded retry limit")

// Func represents functions that can be retried.
type Func func(attempt int) (retry bool, err error)

// Do keeps trying the function until the second argument
// returns false, or no error is returned.
func Do(fn Func) error {
	return DoWithOptions(fn, &Options{DefaultBackoff, MaxRetries})
}

func DoN(fn Func, maxRetries int) error {
	return DoWithOptions(fn, &Options{DefaultBackoff, maxRetries})
}

func DoWithOptions(fn Func, options *Options) error {
	var err error
	var cont bool

	if options.BackoffStrategy == nil {
		options.BackoffStrategy = DefaultBackoff
	}

	if options.MaxRetries == 0 {
		options.MaxRetries = MaxRetries
	}

	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if attempt > options.MaxRetries {
			return errors.Wrap(err, errMaxRetriesReached.Error())
		}

		time.Sleep(options.BackoffStrategy(attempt))
	}
	return err
}

// BackoffStrategy is used to determine how long a retry request should wait until attempted
type BackoffStrategy func(retry int) time.Duration

type Options struct {
	BackoffStrategy BackoffStrategy
	MaxRetries      int
}

// IsMaxRetries checks whether the error is due to hitting the
// maximum number of retries or not.
func IsMaxRetries(err error) bool {
	return err == errMaxRetriesReached
}

// DefaultBackoff always returns 0 seconds
func DefaultBackoff(_ int) time.Duration {
	return 0 * time.Second
}

// ExponentialJitterBackoff returns ever increasing backoffs by a power of 2
// with +/- 0-33% to prevent sychronized reuqests.
func ExponentialJitterBackoff(i int) time.Duration {
	return jitter(int(math.Pow(2, float64(i))))
}

// jitter keeps the +/- 0-33% logic in one place
func jitter(i int) time.Duration {
	ms := i * 1000

	maxJitter := ms / 3

	rand.Seed(time.Now().Unix())
	jitter := rand.Intn(maxJitter + 1)

	if rand.Intn(2) == 1 {
		ms = ms + jitter
	} else {
		ms = ms - jitter
	}

	// a jitter of 0 messes up the time.Tick chan
	if ms <= 0 {
		ms = 1
	}

	return time.Duration(ms) * time.Millisecond
}
