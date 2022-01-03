package services

type mockQueuer struct {
	queueDownloadCallCount int
	queuedIds              []string
}

func (m *mockQueuer) QueueDownload(ids ...string) {
	m.queuedIds = append(m.queuedIds, ids...)
	m.queueDownloadCallCount++
}
