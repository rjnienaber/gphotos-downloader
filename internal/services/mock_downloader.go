package services

import "github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"

type mockDownloader struct {
	get               func(mediaItemId string) (mediaItem models.MediaItem, err error)
	getCallCount      int
	batchGet          func(mediaItemIds []string) (mediaItems models.MediaItemsResult, err error)
	batchGetCallCount int
	list              func(options models.PagingOptions) (mediaItems models.MediaItems, err error)
	search            func(options models.SearchOptions) (mediaItems models.MediaItems, err error)
	download          func(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error)
	downloadCallCount int
}

func (m *mockDownloader) Get(mediaItemId string) (mediaItem models.MediaItem, err error) {
	m.getCallCount++
	if m.get != nil {
		return m.get(mediaItemId)
	}
	return
}

func (m *mockDownloader) BatchGet(mediaItemIds []string) (mediaItems models.MediaItemsResult, err error) {
	m.batchGetCallCount++
	if m.batchGet != nil {
		return m.batchGet(mediaItemIds)
	}
	return
}

func (m *mockDownloader) List(options models.PagingOptions) (mediaItems models.MediaItems, err error) {
	if m.list != nil {
		return m.list(options)
	}
	return
}

func (m *mockDownloader) Search(options models.SearchOptions) (mediaItems models.MediaItems, err error) {
	if m.search != nil {
		return m.search(options)
	}
	return
}

func (m *mockDownloader) Download(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error) {
	m.downloadCallCount++
	if m.download != nil {
		return m.download(tmpDir, baseUrl, isPhoto)
	}
	return
}
