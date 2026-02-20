package reconcileutil

import (
	"crypto/rand"
	"math"
	"math/big"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Requeue returns a ctrl.Result that requeues immediately with the given error (can be nil).
func Requeue(err error) (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, err
}

// RequeueAfter returns a ctrl.Result that requeues after the specified duration with the given error.
func RequeueAfter(d time.Duration, err error) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: d}, err
}

// NoRequeue returns a ctrl.Result with no requeue and the provided error.
func NoRequeue(err error) (ctrl.Result, error) { return ctrl.Result{}, err }

// Backoff computes an exponential backoff with jitter based on retry count.
// base is the initial delay; max caps the maximum delay.
// retry is the attempt number (starting from 0 or 1).
func Backoff(retry int, base, maxDelay time.Duration) time.Duration {
	if retry < 0 {
		retry = 0
	}
	// Exponential growth: base * 2^retry, capped at max
	exp := float64(base) * math.Pow(2, float64(retry))
	d := min(time.Duration(exp), maxDelay)
	// Full jitter: [0, d]
	if d <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(d)+1))
	if err != nil {
		// Fallback to no jitter on failure
		return d
	}
	return time.Duration(n.Int64())
}

// RequeueBackoff returns a ctrl.Result after an exponential backoff with jitter.
func RequeueBackoff(retry int, base, maxDelay time.Duration, err error) (ctrl.Result, error) {
	return RequeueAfter(Backoff(retry, base, maxDelay), err)
}
