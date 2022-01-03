package services

type NoRetryFactory struct {
}

var noRetryTracker NoRetryTracker

func (n NoRetryFactory) Create() RetryTracker {
	return noRetryTracker
}

type NoRetryTracker struct {
}

func (n NoRetryTracker) ShouldRetry(_ error) bool {
	return false
}

func (n NoRetryTracker) Wait() {
}
