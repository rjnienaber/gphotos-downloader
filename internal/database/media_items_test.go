package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaveAndRetrieveMediaItem(t *testing.T) {
	mediaItem := CreateTestMediaItem(t)
	db := CreateTestDatabase(t)

	err := db.MediaItems.Save(&mediaItem)
	assert.NoError(t, err)
	assert.NotNil(t, mediaItem.Uuid)

	dbMediaItem, err := db.MediaItems.Get(mediaItem.Uuid)
	assert.NoError(t, err)
	assert.EqualValues(t, mediaItem, dbMediaItem)
}

func TestMarkMediaItemAsSynced(t *testing.T) {
	mediaItem := CreateTestMediaItem(t)
	mediaItem.Downloaded = false
	mediaItem.FileSize = 0
	mediaItem.SyncedAt = time.Time{}
	mediaItem.LastError = "dns error"
	db := CreateTestDatabase(t)

	err := db.MediaItems.Save(&mediaItem)
	assert.NoError(t, err)
	assert.NotNil(t, mediaItem.Uuid)

	now := time.Now()
	err = db.MediaItems.MarkAsSynced(mediaItem.Uuid, 1024)
	assert.NoError(t, err)

	dbMediaItem, err := db.MediaItems.Get(mediaItem.Uuid)
	assert.NoError(t, err)

	assert.Equal(t, true, dbMediaItem.Downloaded)
	assert.Equal(t, 1024, dbMediaItem.FileSize)
	assert.InDelta(t, now.UnixMilli(), dbMediaItem.SyncedAt.UnixMilli(), 10000)
	assert.Empty(t, dbMediaItem.LastError)
}

func TestRetrieveUndownloadedIds(t *testing.T) {
	mediaItem1 := CreateTestMediaItem(t)
	mediaItem2 := CreateTestMediaItem(t)
	mediaItem2.Downloaded = false
	mediaItem3 := CreateTestMediaItem(t)
	mediaItem4 := CreateTestMediaItem(t)
	mediaItem4.Downloaded = false

	db := CreateTestDatabase(t)
	err := db.MediaItems.Save(&mediaItem1, &mediaItem2, &mediaItem3, &mediaItem4)
	assert.NoError(t, err)

	mediaItemIds, err := db.MediaItems.GetNonDownloadedIds()
	assert.NoError(t, err)
	assert.Len(t, mediaItemIds, 2)
	assert.Equal(t, MediaItemIds{Uuid: mediaItem2.Uuid, RemoteId: mediaItem2.RemoteId}, mediaItemIds[0])
	assert.Equal(t, MediaItemIds{Uuid: mediaItem4.Uuid, RemoteId: mediaItem4.RemoteId}, mediaItemIds[1])
}

func TestUpdateBatchUrl(t *testing.T) {
	mediaItem := CreateTestMediaItem(t)
	db := CreateTestDatabase(t)

	err := db.MediaItems.Save(&mediaItem)
	assert.NoError(t, err)
	assert.NotNil(t, mediaItem.Uuid)

	err = db.MediaItems.UpdateBaseUrl(mediaItem.RemoteId, "http://google.com")
	assert.NoError(t, err)

	dbMediaItem, err := db.MediaItems.Get(mediaItem.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, "http://google.com", dbMediaItem.BaseUrl)
}

func TestMediaItemIsPhoto(t *testing.T) {
	mediaItem := CreateTestMediaItem(t)
	assert.True(t, mediaItem.IsPhoto())

	mediaItem.MimeType = ""
	assert.True(t, mediaItem.IsPhoto())

	mediaItem.MimeType = "video/mp4"
	assert.False(t, mediaItem.IsPhoto())
}

func TestMediaItem_MarkAsErrored(t *testing.T) {
	mediaItem := CreateTestMediaItem(t)
	mediaItem.LastError = "network disconnected"
	db := CreateTestDatabase(t)

	err := db.MediaItems.Save(&mediaItem)
	assert.NoError(t, err)

	dbMediaItem, err := db.MediaItems.Get(mediaItem.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, "network disconnected", dbMediaItem.LastError)
}
