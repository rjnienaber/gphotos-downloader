package services

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	api "github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

type SyncService struct {
	api        googlephotos.Downloader
	db         database.PhotoDatabase
	download   DownloaderQueuer
	logger     utils.Logger
	pagingSize int
}

func NewSyncService(api googlephotos.Downloader, db database.PhotoDatabase, download DownloaderQueuer, logger utils.Logger) SyncService {
	return SyncService{api: api, db: db, download: download, logger: logger, pagingSize: 100}
}

func (s *SyncService) Sync() error {
	lastIndex, err := s.db.Settings.LastIndex()
	if err != nil {
		return err
	}
	s.logger.Info.Printf("last update: %s", lastIndex.Format(time.RFC3339))

	now := time.Now()
	if lastIndex == (time.Time{}) {
		err = s.initialIndex()
		if err != nil {
			return err
		}
	} else {
		err := s.findNew(lastIndex)
		if err != nil {
			return err
		}
	}

	err = s.db.Settings.UpdateLastIndex(now)
	if err != nil {
		return err
	}

	return nil
}

func (s *SyncService) findNew(lastIndex time.Time) error {
	endDate := api.SearchDate{Year: 2999, Month: 12, Day: 31}
	options := api.SearchOptions{
		Filters: api.SearchFilters{
			DateFilter: api.SearchDateFilter{
				Ranges: []api.SearchDateRange{
					{
						StartDate: convertToApiSearchDate(lastIndex),
						EndDate:   endDate,
					},
				},
			},
		},
		Size: s.pagingSize,
	}
	for {
		items, err := s.api.Search(options)
		if err != nil {
			return err
		}

		err = s.processItems(items.MediaItems)
		if err != nil {
			return err
		}

		if items.NextPageToken == "" {
			break
		}
		options.Token = items.NextPageToken
	}
	return nil
}

func (s *SyncService) initialIndex() error {
	options := api.PagingOptions{Size: s.pagingSize}
	for {
		items, err := s.api.List(options)
		if err != nil {
			return err
		}

		err = s.processItems(items.MediaItems)
		if err != nil {
			return err
		}

		if items.NextPageToken == "" {
			break
		}
		options.Token = items.NextPageToken
	}
	return nil
}

func (s *SyncService) processItems(items []api.MediaItem) (err error) {
	for _, item := range items {
		dbItem := convertToDatabaseMediaItem(item)[0]
		filename := dbItem.Filename

		for counter := 2; counter < math.MaxInt; counter++ {
			err = s.db.MediaItems.Save(dbItem)
			if err == nil {
				break
			}

			sqliteError, ok := err.(sqlite3.Error)
			if ok &&
				sqliteError.Code == database.ConstraintPrimaryError &&
				sqliteError.ExtendedCode == database.ConstraintUniqueExtendedError {
				if strings.Contains(sqliteError.Error(), "media_items.local_path, media_items.local_filename") {
					newFilename := generateNewFilename(counter, filename)
					dbItem.LocalFilename = newFilename
					s.logger.Debug.Printf("duplicate file '%s' detected in '%s', trying renaming to '%s'", dbItem.LocalFilename, dbItem.LocalPath, newFilename)
					continue
				}
				if strings.Contains(sqliteError.Error(), "media_items.remote_id") {
					s.logger.Info.Printf("remote id '%s' already exists in db, skipping...", item.Id)
					// already exists, so ignore
					err = nil
					goto noqueue
				}
			}

			// unrecognized error so exit all loops
			goto finished
		}

		s.download.QueueDownload(dbItem.Uuid)
	noqueue:
	}
finished:
	return
}

func convertToDatabaseMediaItem(mediaItems ...api.MediaItem) (dbItems []*database.MediaItem) {
	for _, item := range mediaItems {
		createdAt := item.Metadata.CreationTime
		year, month, day := createdAt.Date()
		localPath := strings.Join([]string{
			strconv.Itoa(year),
			strconv.Itoa(int(month)),
			strconv.Itoa(day),
		}, string(os.PathSeparator))

		dbItem := database.MediaItem{
			RemoteId:      item.Id,
			BaseUrl:       item.BaseUrl,
			MimeType:      item.MimeType,
			Filename:      item.Filename,
			Description:   item.Description,
			LocalPath:     localPath,
			LocalFilename: item.Filename,
			CreatedAt:     item.Metadata.CreationTime,
		}

		dbItems = append(dbItems, &dbItem)
	}
	return
}

func convertToApiSearchDate(dateTime time.Time) api.SearchDate {
	year, month, day := dateTime.Date()
	return api.SearchDate{Year: year, Month: int(month), Day: day}
}

func generateNewFilename(counter int, filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("%s_%03d%s", name, counter, ext)
}
