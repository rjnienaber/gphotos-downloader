package services

import (
	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

type UndownloadedService struct {
	api          googlephotos.Downloader
	db           database.PhotoDatabase
	download     DownloaderQueuer
	logger       utils.Logger
	getBatchSize int
}

func NewUndownloadedService(api googlephotos.Downloader, db database.PhotoDatabase, download DownloaderQueuer, logger utils.Logger) UndownloadedService {
	return UndownloadedService{api: api, db: db, download: download, logger: logger, getBatchSize: 50}
}

func (u *UndownloadedService) Update() (err error) {
	mediaItemIds, err := u.db.MediaItems.GetNonDownloadedIds()
	if err != nil {
		return
	}

	if len(mediaItemIds) == 0 {
		return
	}

	remoteIdMapper := map[string]string{}
	for _, ids := range mediaItemIds {
		remoteIdMapper[ids.RemoteId] = ids.Uuid
	}
	chunks := chunkStringArray(mediaItemIds, u.getBatchSize)

	for _, chunk := range chunks {
		var mediaItems models.MediaItemsResult
		mediaItems, err = u.api.BatchGet(chunk)
		if err != nil {
			break
		}

		for _, item := range mediaItems.MediaItems {
			err = u.db.MediaItems.UpdateBaseUrl(item.Id, item.BaseUrl)
			if err != nil {
				goto finished
			}

			u.download.QueueDownload(remoteIdMapper[item.Id])
		}
	}
finished:
	return
}

func chunkStringArray(values []database.MediaItemIds, chunkSize int) (chunks [][]string) {
	for start, end := 0, chunkSize; start < len(values); start, end = start+chunkSize, end+chunkSize {
		if end > len(values) {
			end = len(values)
		}

		var chunk []string
		for i := start; i < end; i++ {
			chunk = append(chunk, values[i].RemoteId)
		}

		chunks = append(chunks, chunk)
	}
	return
}
