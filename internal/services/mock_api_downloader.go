package services

import (
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
)

type mockDownloader struct {
	get               func(mediaItemId string) (mediaItem models.MediaItem, err error)
	getCallCount      int
	batchGet          func(mediaItemIds []string) (mediaItems models.MediaItemsResult, err error)
	batchGetCallCount int
	list              func(options models.PagingOptions) (mediaItems models.MediaItems, err error)
	listCallCount     int
	search            func(options models.SearchOptions) (mediaItems models.MediaItems, err error)
	searchCallCount   int
	download          func(baseUrl string, isPhoto bool) (filePath string, err error)
	downloadCallCount int
}

func (m *mockDownloader) Get(mediaItemId string) (mediaItem models.MediaItem, err error) {
	m.getCallCount++
	if m.get != nil {
		return m.get(mediaItemId)
	}
	panic("implement me")
}

func (m *mockDownloader) BatchGet(mediaItemIds []string) (mediaItems models.MediaItemsResult, err error) {
	m.batchGetCallCount++
	if m.batchGet != nil {
		return m.batchGet(mediaItemIds)
	}
	panic("implement me")
}

func (m *mockDownloader) List(options models.PagingOptions) (mediaItems models.MediaItems, err error) {
	m.listCallCount++
	if m.list != nil {
		return m.list(options)
	}
	panic("implement me")
}

func (m *mockDownloader) Search(options models.SearchOptions) (mediaItems models.MediaItems, err error) {
	m.searchCallCount++
	if m.search != nil {
		return m.search(options)
	}
	panic("implement me")
}

func (m *mockDownloader) Download(baseUrl string, isPhoto bool) (filePath string, err error) {
	m.downloadCallCount++
	if m.download != nil {
		return m.download(baseUrl, isPhoto)
	}
	panic("implement me")
}
