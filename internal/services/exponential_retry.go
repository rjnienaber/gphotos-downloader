package services

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"time"
)

type ExponentialRetryFactory struct {
	baseTimeInSeconds float64
	sideEffects       sideEffects
	maxTimeInSeconds  float64
	errorTypes        []interface{}
}

type sideEffects interface {
	Sleep(d time.Duration)
	RandomFloat64() float64
}

type defaultSideEffects struct {
}

func (d2 defaultSideEffects) RandomFloat64() float64 {
	return rand.Float64()
}

func (d2 defaultSideEffects) Sleep(d time.Duration) {
	time.Sleep(d)
}

func NewExponentialRetryFactory(errorTypes ...interface{}) ExponentialRetryFactory {
	rand.Seed(time.Now().UnixNano())
	return ExponentialRetryFactory{
		baseTimeInSeconds: 1,
		sideEffects:       defaultSideEffects{},
		maxTimeInSeconds:  30,
		errorTypes:        errorTypes,
	}
}

func (e ExponentialRetryFactory) Create() RetryTracker {
	timeout := time.Now().Add(time.Duration(e.maxTimeInSeconds * float64(time.Second)))
	return &ExponentialRetryTracker{
		baseTimeInSeconds: e.baseTimeInSeconds,
		sideEffects:       e.sideEffects,
		timeoutTime:       timeout,
		errorTypes:        e.errorTypes,
	}
}

type ExponentialRetryTracker struct {
	tries             int
	baseTimeInSeconds float64
	sideEffects       sideEffects
	timeoutTime       time.Time
	errorTypes        []interface{}
}

func (e *ExponentialRetryTracker) ShouldRetry(err error) bool {
	e.tries += 1
	if time.Now().After(e.timeoutTime) {
		return false
	}

	if err == nil {
		return true
	}

	if len(e.errorTypes) == 0 {
		return true
	}

	return searchForErrorTypes(e.errorTypes, err)
}

func (e *ExponentialRetryTracker) Wait() {
	// using full jitter algorithm
	// sleep = random_between(0, min(cap, base * 2 ** attempt))
	// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/

	maxSleepTimeInSeconds := e.baseTimeInSeconds * math.Pow(2, float64(e.tries))
	waitTimeInMilliseconds := maxSleepTimeInSeconds * e.sideEffects.RandomFloat64() * 1000
	e.sideEffects.Sleep(time.Duration(int64(waitTimeInMilliseconds)) * time.Millisecond)
}

func searchForErrorTypes(validErrorTypes []interface{}, err interface{}) bool {
	for {
		value := reflect.Indirect(reflect.ValueOf(err))
		if !value.IsValid() {
			break
		}

		if value.Kind() != reflect.Struct {
			break
		}

		for _, errorType := range validErrorTypes {
			if fmt.Sprintf("%T", errorType) == fmt.Sprintf("%T", err) {
				return true
			}
		}

		field := value.FieldByName("Err")
		if !field.IsValid() {
			break
		}
		err = field.Interface()
	}
	return false
}
