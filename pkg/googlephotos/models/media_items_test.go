package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeserializePhotoMediaItemsJson(t *testing.T) {
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
	mediaItems, err := DeserializeMediaItemsJson([]byte(json))
	assert.NoError(t, err)
	assert.Equal(t, "CkgKQnR5cG", mediaItems.NextPageToken)
	assert.Equal(t, json, mediaItems.Raw)
	assert.Len(t, mediaItems.MediaItems, 1)

	mediaItem := mediaItems.MediaItems[0]
	assert.Equal(t, "ALU181g0Vr1nSvTUkldVxUpM7pdR6U", mediaItem.Id)
	assert.Equal(t, "https://photos.google.com/lr/photo/ALU181g0Vr1nSvTUkldVxUpM7pdR6UZUJEVmggRL6nwobskLMw5", mediaItem.ProductUrl)
	assert.Equal(t, "https://lh3.googleusercontent.com/lr/AFBm1_Z49V5pj7o", mediaItem.BaseUrl)
	assert.Equal(t, "image/jpeg", mediaItem.MimeType)
	assert.Equal(t, "Screenshot_20211227-094449_Settings.jpg", mediaItem.Filename)
	assert.Equal(t, "cute photo", mediaItem.Description)

	creationTime, _ := time.Parse(time.RFC3339, "2021-12-27T09:44:49Z")
	assert.Equal(t, creationTime, mediaItem.Metadata.CreationTime)
	assert.Equal(t, "1080", mediaItem.Metadata.Width)
	assert.Equal(t, "2400", mediaItem.Metadata.Height)
	assert.Empty(t, mediaItem.Metadata.Video)
	assert.Equal(t, "Sony", mediaItem.Metadata.Photo.CameraMake)
	assert.Equal(t, "G8441", mediaItem.Metadata.Photo.CameraModel)
	assert.Equal(t, 4.4, mediaItem.Metadata.Photo.FocalLength)
	assert.Equal(t, 2.0, mediaItem.Metadata.Photo.ApertureFNumber)
	assert.Equal(t, 40, mediaItem.Metadata.Photo.IsoEquivalent)
	assert.Equal(t, "0.004999999s", mediaItem.Metadata.Photo.ExposureTime)
}

func TestDeserializeVideoMediaItemsJson(t *testing.T) {
	json := `{
  "mediaItems": [
    {
      "id": "ALU181hKybye51IIjFZIJoQS9fGCHJ4O",
      "productUrl": "https://photos.google.com/lr/photo/ALU181hKybye51IIjFZIJoQS9fGCHJ4O",
      "description": "cute video",
      "baseUrl": "https://lh3.googleusercontent.com/lr/AFBm1_",
      "mimeType": "video/mp4",
      "mediaMetadata": {
        "creationTime": "2021-12-27T17:59:09Z",
        "width": "1920",
        "height": "1080",
        "video": {
          "fps": 60,
          "status": "READY"
        }
      },
      "filename": "20211227_175723.mp4"
    }
  ],
  "nextPageToken": "CkgKQnR5cG"
}`
	mediaItems, err := DeserializeMediaItemsJson([]byte(json))
	assert.NoError(t, err)
	assert.Equal(t, "CkgKQnR5cG", mediaItems.NextPageToken)
	assert.Equal(t, json, mediaItems.Raw)
	assert.Len(t, mediaItems.MediaItems, 1)

	mediaItem := mediaItems.MediaItems[0]
	assert.Equal(t, "ALU181hKybye51IIjFZIJoQS9fGCHJ4O", mediaItem.Id)
	assert.Equal(t, "https://photos.google.com/lr/photo/ALU181hKybye51IIjFZIJoQS9fGCHJ4O", mediaItem.ProductUrl)
	assert.Equal(t, "https://lh3.googleusercontent.com/lr/AFBm1_", mediaItem.BaseUrl)
	assert.Equal(t, "video/mp4", mediaItem.MimeType)
	assert.Equal(t, "20211227_175723.mp4", mediaItem.Filename)
	assert.Equal(t, "cute video", mediaItem.Description)

	creationTime, _ := time.Parse(time.RFC3339, "2021-12-27T17:59:09Z")
	assert.Equal(t, creationTime, mediaItem.Metadata.CreationTime)
	assert.Equal(t, "1920", mediaItem.Metadata.Width)
	assert.Equal(t, "1080", mediaItem.Metadata.Height)
	assert.Empty(t, mediaItem.Metadata.Photo)
	assert.Equal(t, 60.0, mediaItem.Metadata.Video.Fps)
	assert.Equal(t, "READY", mediaItem.Metadata.Video.Status)
}

func TestDeserializeMediaItemJson(t *testing.T) {
	json := `{
  "id": "ALU181gS07lNXbEv",
  "productUrl": "https://photos.google.com/lr/photo/ALU181gS07lNXbEvgT",
  "description": "cute video",
  "baseUrl": "https://lh3.googleusercontent.com/lr/AFBm1_ZHNqYA-PrYEHY8",
  "mimeType": "video/mp4",
  "mediaMetadata": {
    "creationTime": "2021-12-28T16:54:05Z",
    "width": "1920",
    "height": "1080",
    "video": {
      "fps": 60,
      "status": "READY"
    }
  },
  "filename": "20211228_165349.mp4"
}`
	mediaItem, err := DeserializeMediaItemJson([]byte(json))
	assert.NoError(t, err)

	assert.Equal(t, "ALU181gS07lNXbEv", mediaItem.Id)
	assert.Equal(t, "https://photos.google.com/lr/photo/ALU181gS07lNXbEvgT", mediaItem.ProductUrl)
	assert.Equal(t, "https://lh3.googleusercontent.com/lr/AFBm1_ZHNqYA-PrYEHY8", mediaItem.BaseUrl)
	assert.Equal(t, "video/mp4", mediaItem.MimeType)
	assert.Equal(t, "20211228_165349.mp4", mediaItem.Filename)
	assert.Equal(t, "cute video", mediaItem.Description)

	creationTime, _ := time.Parse(time.RFC3339, "2021-12-28T16:54:05Z")
	assert.Equal(t, creationTime, mediaItem.Metadata.CreationTime)
	assert.Equal(t, "1920", mediaItem.Metadata.Width)
	assert.Equal(t, "1080", mediaItem.Metadata.Height)
	assert.Empty(t, mediaItem.Metadata.Photo)
	assert.Equal(t, 60.0, mediaItem.Metadata.Video.Fps)
	assert.Equal(t, "READY", mediaItem.Metadata.Video.Status)
}
