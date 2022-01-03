package services

import (
	"fmt"
	"testing"

	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"github.com/stretchr/testify/assert"
)

// find non downloaded items
//   chunk into groups of 50 and do batch get for new baseUrl
//   update base url for each item
//   send id to downloader (for each item)
func createUndownloadedService(t *testing.T, downloader googlephotos.Downloader, queuer DownloaderQueuer) UndownloadedService {
	db := database.CreateTestDatabase(t)
	return NewUndownloadedService(downloader, db, queuer, db.Logger)
}

func TestUndownloadServiceDoesNothingWhenNoItemsFound(t *testing.T) {
	service := createUndownloadedService(t, nil, nil)
	item := database.CreateTestMediaItem(t)
	err := service.db.MediaItems.Save(&item)
	assert.NoError(t, err)

	err = service.Update()
	assert.NoError(t, err)
}

func TestUndownloadServiceUpdatesBatchUrlForUndownloadedItems(t *testing.T) {
	_, items := models.CreateTestMediaItemsResult(t)
	downloader := mockDownloader{
		batchGet: func(ids []string) (mediaItems models.MediaItemsResult, err error) {
			assert.Equal(t, []string{items.MediaItems[0].Id}, ids)
			return items, nil
		},
	}

	queuer := mockQueuer{}
	service := createUndownloadedService(t, &downloader, &queuer)
	undownloadedItem := database.CreateTestMediaItem(t)
	undownloadedItem.RemoteId = items.MediaItems[0].Id
	undownloadedItem.Downloaded = false
	downloadedItem := database.CreateTestMediaItem(t)

	err := service.db.MediaItems.Save(&undownloadedItem, &downloadedItem)
	assert.NoError(t, err)

	err = service.Update()
	assert.NoError(t, err)

	assert.Equal(t, 1, downloader.batchGetCallCount)
	assert.Equal(t, 1, queuer.queueDownloadCallCount)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 2)

	dbUndownloadedItem, err := service.db.MediaItems.Get(undownloadedItem.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, items.MediaItems[0].BaseUrl, dbUndownloadedItem.BaseUrl)

	dbDownloadedItem, err := service.db.MediaItems.Get(downloadedItem.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, downloadedItem.BaseUrl, dbDownloadedItem.BaseUrl)
	assert.Equal(t, []string{dbUndownloadedItem.Uuid}, queuer.queuedIds)
}

func TestUndownloadedServiceMakesCallsInGroups(t *testing.T) {
	_, items := models.CreateTestMediaItemsResult(t)
	_, moreItems := models.CreateTestMediaItemsResult(t)
	moreItems.MediaItems[0].Id = "ALU181hKybye51IIjFZIJoQS9fGCHJ4O"

	downloader := mockDownloader{}
	downloader.batchGet = func(ids []string) (mediaItems models.MediaItemsResult, err error) {
		if downloader.batchGetCallCount == 1 {
			assert.Equal(t, []string{items.MediaItems[0].Id}, ids)
			return items, nil
		} else {
			assert.Equal(t, []string{moreItems.MediaItems[0].Id}, ids)
			return moreItems, nil
		}
	}

	queuer := mockQueuer{}
	service := createUndownloadedService(t, &downloader, &queuer)
	service.getBatchSize = 1

	undownloadedItemOne := database.CreateTestMediaItem(t)
	undownloadedItemOne.RemoteId = items.MediaItems[0].Id
	undownloadedItemOne.Downloaded = false
	undownloadedItemTwo := database.CreateTestMediaItem(t)
	undownloadedItemTwo.RemoteId = moreItems.MediaItems[0].Id
	undownloadedItemTwo.Downloaded = false

	err := service.db.MediaItems.Save(&undownloadedItemOne, &undownloadedItemTwo)
	assert.NoError(t, err)

	err = service.Update()
	assert.NoError(t, err)

	assert.Equal(t, 2, downloader.batchGetCallCount)

	dbItems, err := service.db.MediaItems.GetAll()
	assert.NoError(t, err)
	assert.Len(t, dbItems, 2)

	dbUndownloadedItemOne, err := service.db.MediaItems.Get(undownloadedItemOne.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, items.MediaItems[0].BaseUrl, dbUndownloadedItemOne.BaseUrl)

	dbDownloadedItemTwo, err := service.db.MediaItems.Get(undownloadedItemTwo.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, moreItems.MediaItems[0].BaseUrl, dbDownloadedItemTwo.BaseUrl)
	assert.Equal(t, []string{dbUndownloadedItemOne.Uuid, dbDownloadedItemTwo.Uuid}, queuer.queuedIds)
}

func TestChunkingOfArray(t *testing.T) {
	type testCase struct {
		batchSize int
		values    []database.MediaItemIds
		expected  [][]string
	}

	testCases := []testCase{
		{batchSize: 50, values: []database.MediaItemIds{{RemoteId: "1"}}, expected: [][]string{{"1"}}},
		{batchSize: 1, values: []database.MediaItemIds{{RemoteId: "1"}, {RemoteId: "2"}, {RemoteId: "3"}}, expected: [][]string{{"1"}, {"2"}, {"3"}}},
		{batchSize: 2, values: []database.MediaItemIds{{RemoteId: "1"}, {RemoteId: "2"}, {RemoteId: "3"}}, expected: [][]string{{"1", "2"}, {"3"}}},
		{batchSize: 3, values: []database.MediaItemIds{{RemoteId: "1"}, {RemoteId: "2"}, {RemoteId: "3"}, {RemoteId: "4"}, {RemoteId: "5"}, {RemoteId: "6"}, {RemoteId: "7"}}, expected: [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7"}}},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("batchSize-%d,values-%d", tc.batchSize, len(tc.values)), func(t *testing.T) {
			assert.Equal(t, tc.expected, chunkStringArray(tc.values, tc.batchSize))
		})
	}
}
