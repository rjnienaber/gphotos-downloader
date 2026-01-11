package services

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func createDownloadService(t *testing.T, downloader googlephotos.Downloader, opts ...Option) DownloadService {
	db := database.CreateTestDatabase(t)
	allOptions := append([]Option{WithMaxWorkers(1)}, opts...)
	return NewDownloadService(downloader, db, os.TempDir(), allOptions...)
}

func createMediaItemToDownload(t *testing.T) database.MediaItem {
	item := database.CreateTestMediaItem(t)
	item.Downloaded = false
	item.SyncedAt = time.Time{}
	item.FileSize = 0
	item.LocalFilename = fmt.Sprintf("renamed-downloaded%d-file.tmp", time.Now().UnixNano())
	return item
}

func writeTempFile(t *testing.T, content string) string {
	file, err := ioutil.TempFile("", "downloadedfile.*.tmp")
	assert.NoError(t, err)
	defer utils.CheckClose(file, &err)

	_, err = file.WriteString(content)
	assert.NoError(t, err)

	return file.Name()
}

func assertItemDownloaded(t *testing.T, service DownloadService, id string, finishedTime int64) {
	dbItem, err := service.db.MediaItems.Get(id)
	assert.NoError(t, err)

	assert.True(t, dbItem.Downloaded)
	assert.Equal(t, 4, dbItem.FileSize)
	assert.InDelta(t, finishedTime, dbItem.SyncedAt.UnixMilli(), 10000)
}

func TestDownloadService_DownloadsASingleFile(t *testing.T) {
	item := createMediaItemToDownload(t)

	downloader := mockDownloader{
		download: func(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error) {
			assert.Equal(t, item.BaseUrl, baseUrl)
			assert.Equal(t, item.IsPhoto(), isPhoto)

			return writeTempFile(t, "abcd"), nil
		},
	}

	service := createDownloadService(t, &downloader)
	err := service.db.MediaItems.Save(&item)
	assert.NoError(t, err)

	service.QueueDownload(item.Uuid)
	service.Finish()
	finishedTime := time.Now().UnixMilli()

	assert.Equal(t, 1, downloader.downloadCallCount)
	assertItemDownloaded(t, service, item.Uuid, finishedTime)
}

func TestDownloadService_DownloadsMultipleFiles(t *testing.T) {
	itemOne := createMediaItemToDownload(t)
	itemTwo := createMediaItemToDownload(t)

	downloader := mockDownloader{
		download: func(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error) {
			if itemOne.BaseUrl == baseUrl {
				assert.Equal(t, itemOne.BaseUrl, baseUrl)
				assert.Equal(t, itemOne.IsPhoto(), isPhoto)
			}

			if itemTwo.BaseUrl == baseUrl {
				assert.Equal(t, itemTwo.BaseUrl, baseUrl)
				assert.Equal(t, itemTwo.IsPhoto(), isPhoto)
			}

			return writeTempFile(t, "abcd"), nil
		},
	}

	service := createDownloadService(t, &downloader)
	err := service.db.MediaItems.Save(&itemOne, &itemTwo)
	assert.NoError(t, err)

	service.QueueDownload(itemOne.Uuid, itemTwo.Uuid)
	service.Finish()
	finishedTime := time.Now().UnixMilli()

	assert.Equal(t, 2, downloader.downloadCallCount)
	for _, id := range []string{itemOne.Uuid, itemTwo.Uuid} {
		assertItemDownloaded(t, service, id, finishedTime)
	}
}

func TestDownloadService_UpdatesBaseUrlIfIncorrect(t *testing.T) {
	item := createMediaItemToDownload(t)
	newBaseUrl := "https://lh3.googleusercontent.com/lr/AFBm1_bKC3xpsBsbtwcD3wKVcEMdwlf0Sk61"
	downloader := mockDownloader{}
	downloader.download = func(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error) {
		if downloader.downloadCallCount == 1 {
			assert.Equal(t, item.BaseUrl, baseUrl)
			assert.Equal(t, item.IsPhoto(), isPhoto)
			return "", models.ApiError{StatusCode: 403}
		}
		if downloader.downloadCallCount == 2 {
			assert.Equal(t, newBaseUrl, baseUrl)
			assert.Equal(t, item.IsPhoto(), isPhoto)
		}

		return writeTempFile(t, "abcd"), nil
	}
	downloader.get = func(mediaItemId string) (mediaItem models.MediaItem, err error) {
		assert.Equal(t, item.RemoteId, mediaItemId)
		apiItem := createMediaItem(t)
		apiItem.BaseUrl = newBaseUrl
		return apiItem, nil
	}

	service := createDownloadService(t, &downloader)
	err := service.db.MediaItems.Save(&item)
	assert.NoError(t, err)

	service.QueueDownload(item.Uuid)
	service.Finish()
	finishedTime := time.Now().UnixMilli()

	assert.Equal(t, 2, downloader.downloadCallCount)
	assert.Equal(t, 1, downloader.getCallCount)

	dbItem, err := service.db.MediaItems.Get(item.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, newBaseUrl, dbItem.BaseUrl)

	assertItemDownloaded(t, service, item.Uuid, finishedTime)
}

func TestDownloadService_HandlesNetworkFailureAndRetries(t *testing.T) {
	item := createMediaItemToDownload(t)
	downloader := mockDownloader{}
	downloader.download = func(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error) {
		if downloader.downloadCallCount < 3 {
			return "", models.ApiError{StatusCode: 500}
		}
		assert.Equal(t, item.BaseUrl, baseUrl)
		assert.Equal(t, item.IsPhoto(), isPhoto)

		return writeTempFile(t, "abcd"), nil
	}

	factory := NewExponentialRetryFactory()
	factory.baseTimeInSeconds = 0.1
	service := createDownloadService(t, &downloader, WithRetryFactory(factory))
	err := service.db.MediaItems.Save(&item)
	assert.NoError(t, err)

	service.QueueDownload(item.Uuid)
	service.Finish()
	finishedTime := time.Now().UnixMilli()

	assert.Equal(t, 3, downloader.downloadCallCount)
	assertItemDownloaded(t, service, item.Uuid, finishedTime)
}

func TestDownloadService_LogErrorInDatabase(t *testing.T) {
	item := createMediaItemToDownload(t)

	downloader := mockDownloader{
		download: func(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error) {
			return "", errors.New("invalid url")
		},
	}

	service := createDownloadService(t, &downloader)
	err := service.db.MediaItems.Save(&item)
	assert.NoError(t, err)

	service.QueueDownload(item.Uuid)
	service.Finish()

	assert.Equal(t, 1, downloader.downloadCallCount)
	dbItem, err := service.db.MediaItems.Get(item.Uuid)
	assert.NoError(t, err)

	assert.False(t, dbItem.Downloaded)
	assert.Equal(t, 0, dbItem.FileSize)
	assert.Empty(t, dbItem.SyncedAt)
	assert.Equal(t, "invalid url", dbItem.LastError)
}

func TestDownloadService_CheckItemHasNotBeenDownloaded(t *testing.T) {
	item := database.CreateTestMediaItem(t)
	service := createDownloadService(t, nil)
	err := service.db.MediaItems.Save(&item)
	assert.NoError(t, err)

	service.QueueDownload(item.Uuid)
	service.Finish()

	dbItem, err := service.db.MediaItems.Get(item.Uuid)
	assert.NoError(t, err)

	assert.True(t, dbItem.Downloaded)
	assert.Equal(t, 2345, dbItem.FileSize)
	assert.NotEmpty(t, dbItem.SyncedAt)
	assert.Empty(t, dbItem.LastError)
}
