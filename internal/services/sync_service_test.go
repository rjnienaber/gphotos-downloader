package services

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"github.com/stretchr/testify/assert"
)

func TestConvertApiMediaItemToDatabaseMediaItem(t *testing.T) {
	apiMediaItems := createMediaItems(t)

	dbMediaItems := convertToDatabaseMediaItem(apiMediaItems.MediaItems...)

	dbMediaItem := dbMediaItems[0]
	assert.Empty(t, dbMediaItem.Uuid)
	assert.Equal(t, "ALU181g0Vr1nSvTUkldVxUpM7pdR6U", dbMediaItem.RemoteId)
	assert.Equal(t, "https://lh3.googleusercontent.com/lr/AFBm1_Z49V5pj7o", dbMediaItem.BaseUrl)
	assert.Equal(t, "image/jpeg", dbMediaItem.MimeType)
	assert.Equal(t, "Screenshot_20211227-094449_Settings.jpg", dbMediaItem.Filename)
	assert.Equal(t, "cute photo", dbMediaItem.Description)
	assert.False(t, dbMediaItem.Downloaded)

	localPath := filepath.Join("2021", "12", "27")
	assert.Equal(t, localPath, dbMediaItem.LocalPath)
	assert.Equal(t, "Screenshot_20211227-094449_Settings.jpg", dbMediaItem.LocalFilename)
	assert.Equal(t, 0, dbMediaItem.FileSize)

	creationTime, _ := time.Parse(time.RFC3339, "2021-12-27T09:44:49Z")
	assert.Equal(t, creationTime, dbMediaItem.CreatedAt)
	assert.Equal(t, time.Time{}, dbMediaItem.ModifiedAt)
	assert.Equal(t, time.Time{}, dbMediaItem.SyncedAt)
}

func createMediaItems(t *testing.T) models.MediaItems {
	json := `{
	 "mediaItems": [
	   {
	     "id": "ALU181g0Vr1nSvTUkldVxUpM7pdR6U",
	     "productUrl": "https://photos.google.com/lr/photo/ALU181g0Vr1nSvTUkldVxUpM7pdR6UZUJEVmggRL6nwobskLMw5",
		 "description": "cute photo",
		 "baseUrl": "https://lh3.googleusercontent.com/lr/AFBm1_Z49V5pj7o",
	     "mimeType": "image/jpeg",
	     "mediaMetadata": {
	       "creationTime": "2021-12-27T09:44:49Z",
	       "width": "1080",
	       "height": "2400",
	       "photo": {
	         "cameraMake": "Sony",
	         "cameraModel": "G8441",
	         "focalLength": 4.4,
	         "apertureFNumber": 2,
	         "isoEquivalent": 40,
	         "exposureTime": "0.004999999s"
	       }
	     },
	     "filename": "Screenshot_20211227-094449_Settings.jpg"
	   }
	 ],
	 "nextPageToken": "CkgKQnR5cG"
	}`

	apiMediaItems, err := models.DeserializeMediaItemsJson([]byte(json))
	assert.NoError(t, err)
	assert.Len(t, apiMediaItems.MediaItems, 1)
	return apiMediaItems
}

func createMediaItem(t *testing.T) models.MediaItem {
	return createMediaItems(t).MediaItems[0]
}

func createSyncService(t *testing.T, downloader googlephotos.Downloader, queuer DownloaderQueuer) SyncService {
	db := database.CreateTestDatabase(t)
	return NewSyncService(downloader, db, queuer, db.Logger)
}

func TestSyncServiceSyncOnInitialIndex(t *testing.T) {
	items := createMediaItems(t)
	items.NextPageToken = ""
	downloader := mockDownloader{
		list: func(options models.PagingOptions) (mediaItems models.MediaItems, err error) {
			return items, nil
		},
	}

	queuer := mockQueuer{}
	service := createSyncService(t, &downloader, &queuer)

	err := service.Sync()
	assert.NoError(t, err)

	lastIndex, err := service.db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.InDelta(t, time.Now().UnixMilli(), lastIndex.UnixMilli(), 10000)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 1)

	dbItem := dbItems[0]
	assert.Equal(t, items.MediaItems[0].Id, dbItem.RemoteId)
	assert.False(t, dbItem.Downloaded)
	assert.Equal(t, []string{dbItem.Uuid}, queuer.queuedIds)
}

func TestSyncServiceSyncOnInitialIndexPaging(t *testing.T) {
	firstItems := createMediaItems(t)
	secondItems := createMediaItems(t)
	secondItems.MediaItems[0].Id = "ALU181gS07lNXbEvg"
	secondItems.MediaItems[0].Filename = "Screenshot_2.jpg"
	secondItems.NextPageToken = ""

	pagingTest := true
	downloader := mockDownloader{
		list: func(_ models.PagingOptions) (mediaItems models.MediaItems, err error) {
			if pagingTest {
				pagingTest = false
				return firstItems, nil
			} else {
				return secondItems, nil
			}
		},
	}

	queuer := mockQueuer{}
	service := createSyncService(t, &downloader, &queuer)

	err := service.Sync()
	assert.NoError(t, err)

	lastIndex, err := service.db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.InDelta(t, time.Now().UnixMilli(), lastIndex.UnixMilli(), 10000)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 2)

	dbItemOne := dbItems[0]
	assert.Equal(t, firstItems.MediaItems[0].Id, dbItemOne.RemoteId)
	assert.False(t, dbItemOne.Downloaded)

	dbItemTwo := dbItems[1]
	assert.Equal(t, secondItems.MediaItems[0].Id, dbItemTwo.RemoteId)
	assert.False(t, dbItemTwo.Downloaded)

	assert.Equal(t, []string{dbItemOne.Uuid, dbItemTwo.Uuid}, queuer.queuedIds)
}

func TestSyncServiceSyncSearchForUpdates(t *testing.T) {
	items := createMediaItems(t)
	items.NextPageToken = ""
	downloader := mockDownloader{
		search: func(_ models.SearchOptions) (mediaItems models.MediaItems, err error) {
			return items, nil
		},
	}

	queuer := mockQueuer{}
	service := createSyncService(t, &downloader, &queuer)
	err := service.db.Settings.UpdateLastIndex(time.Now())
	assert.NoError(t, err)

	err = service.Sync()
	assert.NoError(t, err)

	lastIndex, err := service.db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.InDelta(t, time.Now().UnixMilli(), lastIndex.UnixMilli(), 10000)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 1)

	dbItem := dbItems[0]
	assert.Equal(t, items.MediaItems[0].Id, dbItem.RemoteId)
	assert.False(t, dbItem.Downloaded)
	assert.Equal(t, []string{dbItem.Uuid}, queuer.queuedIds)
}

func TestSyncServiceSyncSearchForUpdatesPaging(t *testing.T) {
	firstItems := createMediaItems(t)
	secondItems := createMediaItems(t)
	secondItems.MediaItems[0].Id = "ALU181gS07lNXbEvg"
	secondItems.MediaItems[0].Filename = "Screenshot_2.jpg"
	secondItems.NextPageToken = ""

	pagingTest := true
	downloader := mockDownloader{
		search: func(_ models.SearchOptions) (mediaItems models.MediaItems, err error) {
			if pagingTest {
				pagingTest = false
				return firstItems, nil
			} else {
				return secondItems, nil
			}
		},
	}

	queuer := mockQueuer{}
	service := createSyncService(t, &downloader, &queuer)
	err := service.db.Settings.UpdateLastIndex(time.Now())
	assert.NoError(t, err)

	err = service.Sync()
	assert.NoError(t, err)

	lastIndex, err := service.db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.InDelta(t, time.Now().UnixMilli(), lastIndex.UnixMilli(), 10000)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 2)

	dbItemOne := dbItems[0]
	assert.Equal(t, firstItems.MediaItems[0].Id, dbItemOne.RemoteId)
	assert.False(t, dbItemOne.Downloaded)

	dbItemTwo := dbItems[1]
	assert.Equal(t, secondItems.MediaItems[0].Id, dbItemTwo.RemoteId)
	assert.False(t, dbItemTwo.Downloaded)
	assert.Equal(t, []string{dbItemOne.Uuid, dbItemTwo.Uuid}, queuer.queuedIds)
}

func TestHandlesLocalPathDuplicates(t *testing.T) {
	items := createMediaItems(t)
	items.NextPageToken = ""
	duplicateFilenameItem := createMediaItem(t)
	duplicateFilenameItem.Id = "ALU181gS07lNXbEvg"
	items.MediaItems = append(items.MediaItems, duplicateFilenameItem)

	downloader := mockDownloader{
		list: func(_ models.PagingOptions) (mediaItems models.MediaItems, err error) {
			return items, nil
		},
	}

	queuer := mockQueuer{}
	service := createSyncService(t, &downloader, &queuer)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 0)

	err = service.Sync()
	assert.NoError(t, err)

	dbItems, err = service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 2)

	dbItemOne := dbItems[0]
	assert.Equal(t, items.MediaItems[0].Id, dbItemOne.RemoteId)
	assert.Equal(t, "Screenshot_20211227-094449_Settings.jpg", dbItemOne.Filename)
	assert.Equal(t, "Screenshot_20211227-094449_Settings.jpg", dbItemOne.LocalFilename)

	dbItemTwo := dbItems[1]
	assert.Equal(t, items.MediaItems[1].Id, dbItemTwo.RemoteId)
	assert.Equal(t, "Screenshot_20211227-094449_Settings.jpg", dbItemTwo.Filename)
	assert.Equal(t, "Screenshot_20211227-094449_Settings_002.jpg", dbItemTwo.LocalFilename)
	assert.Equal(t, []string{dbItemOne.Uuid, dbItemTwo.Uuid}, queuer.queuedIds)
}

func TestAppendingNumberedPrefixesToFiles(t *testing.T) {
	type testCase struct {
		counter          int
		filename         string
		expectedFilename string
	}
	testCases := []testCase{
		{counter: 1, filename: "test.jpg", expectedFilename: "test_001.jpg"},
		{counter: 20, filename: "test", expectedFilename: "test_020"},
		{counter: 1, filename: "test_001.jpg", expectedFilename: "test_001_001.jpg"},
		{counter: 1, filename: "test.jpg.jpg", expectedFilename: "test.jpg_001.jpg"},
		{counter: 1, filename: "test_001.jpg.jpg", expectedFilename: "test_001.jpg_001.jpg"},
		{counter: 999, filename: "test.jpg", expectedFilename: "test_999.jpg"},
		{counter: 1000, filename: "test.jpg", expectedFilename: "test_1000.jpg"},
		{counter: 1001, filename: "test.jpg", expectedFilename: "test_1001.jpg"},
		{counter: 10000, filename: "test.jpg", expectedFilename: "test_10000.jpg"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s -> %s", tc.filename, tc.expectedFilename), func(t *testing.T) {
			assert.Equal(t, tc.expectedFilename, generateNewFilename(tc.counter, tc.filename))
		})
	}
}

func TestHandlesFilesAlreadyIndexed(t *testing.T) {
	items := createMediaItems(t)
	items.MediaItems = append(items.MediaItems, createMediaItem(t))
	items.NextPageToken = ""

	downloader := mockDownloader{
		list: func(_ models.PagingOptions) (mediaItems models.MediaItems, err error) {
			return items, nil
		},
	}

	queuer := mockQueuer{}
	service := createSyncService(t, &downloader, &queuer)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 0)

	err = service.Sync()
	assert.NoError(t, err)

	dbItems, err = service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 1)

	dbItemOne := dbItems[0]
	assert.Equal(t, items.MediaItems[0].Id, dbItemOne.RemoteId)
	assert.Equal(t, []string{dbItemOne.Uuid}, queuer.queuedIds)
}
