package models

import (
	json2 "encoding/json"
	"time"
)

type PagingOptions struct {
	Size  int    `json:"pageSize,omitempty"`
	Token string `json:"pageToken,omitempty"`
}

type MediaItems struct {
	MediaItems    []MediaItem `json:"mediaItems"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
	Raw           string
}

type MediaItem struct {
	Id          string        `json:"id"`
	Description string        `json:"description,omitempty"`
	ProductUrl  string        `json:"productUrl"`
	BaseUrl     string        `json:"baseUrl"`
	MimeType    string        `json:"mimeType"`
	Metadata    MediaMetadata `json:"mediaMetadata"`
	Filename    string        `json:"filename"`
}

type MediaMetadata struct {
	CreationTime time.Time      `json:"creationTime"`
	Width        string         `json:"width"`
	Height       string         `json:"height"`
	Photo        MediaItemPhoto `json:"photo,omitempty"`
	Video        MediaItemVideo `json:"video,omitempty"`
}

type MediaItemPhoto struct {
	CameraMake      string  `json:"cameraMake,omitempty"`
	CameraModel     string  `json:"cameraModel,omitempty"`
	FocalLength     float64 `json:"focalLength,omitempty"`
	ApertureFNumber float64 `json:"apertureFNumber,omitempty"`
	IsoEquivalent   int     `json:"isoEquivalent,omitempty"`
	ExposureTime    string  `json:"exposureTime,omitempty"`
}

type MediaItemVideo struct {
	Fps    float64 `json:"fps,omitempty"`
	Status string  `json:"status,omitempty"`
}

func DeserializeMediaItemsJson(body []byte) (mediaItems MediaItems, err error) {
	mediaItems.Raw = string(body)
	err = json2.Unmarshal(body, &mediaItems)
	return
}

func DeserializeMediaItemJson(body []byte) (mediaItem MediaItem, err error) {
	err = json2.Unmarshal(body, &mediaItem)
	return
}
