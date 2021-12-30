package models

import (
	json2 "encoding/json"
)

type mediaItemResult struct {
	MediaItem MediaItem `json:"mediaItem"`
}

type mediaItemsResult struct {
	MediaItemResults []json2.RawMessage `json:"mediaItemResults"`
}

type ErrorResult struct {
	Id     string
	Status ErrorStatus `json:"status"`
}

type ErrorStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MediaItemsResult struct {
	MediaItems []MediaItem
	Errors     []ErrorResult
	Raw        string
}

func DeserializeMediaItemsResultJson(body []byte, ids []string) (results MediaItemsResult, err error) {
	var data mediaItemsResult
	err = json2.Unmarshal(body, &data)
	if err != nil {
		return
	}

	var mediaItems []MediaItem
	var errors []ErrorResult
	for i, result := range data.MediaItemResults {
		var item mediaItemResult
		err = json2.Unmarshal(result, &item)
		if item != (mediaItemResult{}) {
			mediaItems = append(mediaItems, item.MediaItem)
			continue
		} else if err != nil {
			return
		}

		var errorResult ErrorResult
		err = json2.Unmarshal(result, &errorResult)
		if errorResult != (ErrorResult{}) {
			errorResult.Id = ids[i]
			errors = append(errors, errorResult)
		} else if err != nil {
			return
		}
	}

	results.MediaItems = mediaItems
	results.Errors = errors
	results.Raw = string(body)

	return
}
