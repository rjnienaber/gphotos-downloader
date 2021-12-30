package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func CreateTestMediaItemsResult(t *testing.T) (string, MediaItemsResult) {
	json := `{
  "mediaItemResults": [
    {
      "mediaItem": {
        "id": "ALU181gS07lNXbEvg",
        "productUrl": "https://photos.google.com/lr/photo/ALU181gS07lNXbEvgT9RFPesC0Kvx",
        "baseUrl": "https://lh3.googleusercontent.com/lr/AFBm1_bKC3xpsBsbtwcD3wKVcEMdwlf0Sk61",
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
      }
    },
	{
      "id": "bPQKpSPtCtPnpn3eqZLWJ9Jo",
      "status": {
        "code": 3,
        "message": "Invalid media item ID."
      }
    }
  ]
}`

	ids := []string{"ALU181gS07lNXbEvg", "bPQKpSPtCtPnpn3eqZLWJ9Jo"}
	mediaItems, err := DeserializeMediaItemsResultJson([]byte(json), ids)
	assert.NoError(t, err)
	return json, mediaItems
}
