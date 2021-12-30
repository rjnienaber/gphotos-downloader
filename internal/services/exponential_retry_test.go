package services

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockSleeper struct {
	durations []int64
}

func (m *mockSleeper) RandomFloat64() float64 {
	return 1
}

func (m *mockSleeper) Sleep(d time.Duration) {
	m.durations = append(m.durations, d.Milliseconds())
}

func TestExponentialRetryTracker_ShouldRetryUntilTimeoutReached(t *testing.T) {
	factory := ExponentialRetryFactory{
		baseTimeInSeconds: 1,
		sideEffects:       &mockSleeper{},
		maxTimeInSeconds:  30,
	}

	tracker, ok := factory.Create().(*ExponentialRetryTracker)
	assert.True(t, ok)

	tracker.timeoutTime = time.Now()
	assert.False(t, tracker.ShouldRetry(errors.New("test")))
}

func TestExponentialRetryTracker_Wait_DurationsAreCorrect(t *testing.T) {
	sleeper := &mockSleeper{}
	factory := ExponentialRetryFactory{
		baseTimeInSeconds: 1,
		sideEffects:       sleeper,
		maxTimeInSeconds:  30,
	}

	tracker, ok := factory.Create().(*ExponentialRetryTracker)
	assert.True(t, ok)

	for i := 0; i < 4; i++ {
		tracker.ShouldRetry(errors.New("wait test"))
		tracker.Wait()
	}

	assert.Equal(t, []int64{2000, 4000, 8000, 16000}, sleeper.durations)
}

func TestExponentialRetryTracker_Wait_UnevenDurationsAreCorrect(t *testing.T) {
	sleeper := &mockSleeper{}
	factory := ExponentialRetryFactory{
		baseTimeInSeconds: 0.3,
		sideEffects:       sleeper,
		maxTimeInSeconds:  30,
	}

	tracker, ok := factory.Create().(*ExponentialRetryTracker)
	assert.True(t, ok)

	for i := 0; i < 4; i++ {
		tracker.ShouldRetry(errors.New("wait test"))
		tracker.Wait()
	}

	assert.Equal(t, []int64{600, 1200, 2400, 4800}, sleeper.durations)
}

func TestExponentialRetryTracker_ShouldRetryReturnsTrueOnCertainErrors(t *testing.T) {
	acceptableError := &net.OpError{
		Op:     "dial",
		Net:    "tcp",
		Source: nil,
		Addr:   nil,
		Err:    nil,
	}

	acceptableNestedError := url.Error{
		Op:  "Get",
		URL: "http://blah.blah.com/",
		Err: &url.Error{
			Op:  "NestedGet",
			URL: "http://nested.url.com/",
			Err: acceptableError,
		},
	}

	notAcceptableNestedError := url.Error{
		Op:  "Get",
		URL: "http://blah.blah.com/",
		Err: &url.Error{
			Op:  "NestedGet",
			URL: "http://nested.url.com/",
			Err: nil,
		},
	}

	type testCase struct {
		err  error
		pass bool
	}

	testCases := []testCase{
		{err: nil, pass: true},
		{err: errors.New("misc error"), pass: false},
		{err: acceptableError, pass: true},
		{err: &acceptableNestedError, pass: true},
		{err: &notAcceptableNestedError, pass: false},
	}

	factory := ExponentialRetryFactory{
		baseTimeInSeconds: 1,
		sideEffects:       &mockSleeper{},
		maxTimeInSeconds:  30,
		errorTypes: []interface{}{
			net.OpError{},
			new(net.OpError),
		},
	}
	tracker, ok := factory.Create().(*ExponentialRetryTracker)
	assert.True(t, ok)

	for i, tc := range testCases {
		t.Run("test case "+strconv.Itoa(i+1), func(t *testing.T) {
			assert.Equal(t, tc.pass, tracker.ShouldRetry(tc.err))
		})
	}
}

func TestExponentialRetryTracker_ShouldRetry_ReturnsTrueIfNoErrorsSpecified(t *testing.T) {
	factory := ExponentialRetryFactory{
		baseTimeInSeconds: 1000,
		sideEffects:       &mockSleeper{},
		maxTimeInSeconds:  30,
	}
	tracker, ok := factory.Create().(*ExponentialRetryTracker)
	assert.True(t, ok)

	assert.True(t, tracker.ShouldRetry(errors.New("test error")))
}
