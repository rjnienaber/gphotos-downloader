package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeserializeMediaItemsResultJson(t *testing.T) {
	json, mediaItems := CreateTestMediaItemsResult(t)
	assert.Equal(t, json, mediaItems.Raw)
	assert.Len(t, mediaItems.MediaItems, 1)

	mediaItem := mediaItems.MediaItems[0]
	assert.Equal(t, "ALU181gS07lNXbEvg", mediaItem.Id)
	assert.Equal(t, "https://photos.google.com/lr/photo/ALU181gS07lNXbEvgT9RFPesC0Kvx", mediaItem.ProductUrl)
	assert.Equal(t, "https://lh3.googleusercontent.com/lr/AFBm1_bKC3xpsBsbtwcD3wKVcEMdwlf0Sk61", mediaItem.BaseUrl)
	assert.Equal(t, "video/mp4", mediaItem.MimeType)
	assert.Equal(t, "20211228_165349.mp4", mediaItem.Filename)

	creationTime, _ := time.Parse(time.RFC3339, "2021-12-28T16:54:05Z")
	assert.Equal(t, creationTime, mediaItem.Metadata.CreationTime)
	assert.Equal(t, "1920", mediaItem.Metadata.Width)
	assert.Equal(t, "1080", mediaItem.Metadata.Height)
	assert.Empty(t, mediaItem.Metadata.Photo)
	assert.Equal(t, 60.0, mediaItem.Metadata.Video.Fps)
	assert.Equal(t, "READY", mediaItem.Metadata.Video.Status)

	assert.Len(t, mediaItems.Errors, 1)

	retrieveError := mediaItems.Errors[0]
	assert.Equal(t, "bPQKpSPtCtPnpn3eqZLWJ9Jo", retrieveError.Id)

	assert.Equal(t, 3, retrieveError.Status.Code)
	assert.Equal(t, "Invalid media item ID.", retrieveError.Status.Message)

}
