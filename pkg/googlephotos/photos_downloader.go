package googlephotos

import "github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"

type Downloader interface {
	Get(mediaItemId string) (mediaItem models.MediaItem, err error)
	BatchGet(mediaItemIds []string) (mediaItems models.MediaItemsResult, err error)
	List(options models.PagingOptions) (mediaItems models.MediaItems, err error)
	Search(options models.SearchOptions) (mediaItems models.MediaItems, err error)
	Download(baseUrl string, isPhoto bool) (filePath string, err error)
}
