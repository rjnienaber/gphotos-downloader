package database

import (
	"database/sql"
	"strings"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

type mediaItems struct {
	sqlFuncs SqlFuncs
	logger   utils.Logger
}

type MediaItem struct {
	Uuid          string
	RemoteId      string
	BaseUrl       string
	MimeType      string
	Filename      string
	Description   string
	Downloaded    bool
	LocalPath     string
	LocalFilename string
	FileSize      int
	CreatedAt     time.Time
	ModifiedAt    time.Time
	SyncedAt      time.Time
	LastError     string
}

type MediaItemIds struct {
	Uuid     string
	RemoteId string
}

func (m MediaItem) IsPhoto() bool {
	return !strings.Contains(m.MimeType, "video")
}

func (m *mediaItems) Save(items ...*MediaItem) error {
	var values []string
	var params []interface{}
	for _, item := range items {
		if item.Uuid == "" {
			newUUid, err := uuid.NewV4()
			if err != nil {
				return err
			}
			item.Uuid = newUUid.String()
		}

		params = append(params, item.Uuid, item.RemoteId, item.BaseUrl, item.MimeType, item.Filename)
		params = append(params, item.Description, item.Downloaded, item.LocalPath, item.LocalFilename, item.FileSize)
		params = append(params, item.CreatedAt.Format(time.RFC3339Nano), item.ModifiedAt.Format(time.RFC3339Nano))
		params = append(params, item.SyncedAt.Format(time.RFC3339Nano), item.LastError)
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	}

	insertSql := "INSERT INTO media_items VALUES" + strings.Join(values, ", ")
	err := m.sqlFuncs.Exec(insertSql, params...)

	return err
}

func (m *mediaItems) Get(id string) (mediaItem MediaItem, err error) {
	query := `SELECT uuid, remote_id, base_url, mime_type, filename, description, downloaded, 
					 local_path, local_filename, file_size, created_at, modified_at, synced_at, last_error 
			  FROM media_items WHERE uuid = ?`
	args := []interface{}{id}

	err = m.sqlFuncs.QueryRow(query, args, mediaItemRowMapper(&mediaItem))

	if err != nil {
		return MediaItem{}, err
	}

	return
}

func (m *mediaItems) MarkAsSynced(id string, fileSize int64) error {
	updateSql := "UPDATE media_items SET downloaded = ?, file_size = ?, synced_at = ?, last_error = '' WHERE uuid = ?"
	return m.sqlFuncs.Exec(updateSql, true, fileSize, time.Now().Format(time.RFC3339Nano), id)
}

func (m *mediaItems) MarkAsErrored(id string, err error) error {
	updateSql := "UPDATE media_items SET last_error = ? WHERE uuid = ?"
	return m.sqlFuncs.Exec(updateSql, err.Error(), id)
}

func (m *mediaItems) GetNonDownloadedIds() (mediaItemIds []MediaItemIds, err error) {
	selectSql := "SELECT uuid, remote_id FROM media_items WHERE downloaded = 0"

	mapper := func(row Scanner) (mapperError error) {
		var ids MediaItemIds
		mapperError = row.Scan(&ids.Uuid, &ids.RemoteId)
		if mapperError == nil {
			mediaItemIds = append(mediaItemIds, ids)
		}
		return
	}

	err = m.sqlFuncs.Query(mapper, selectSql)
	if err != nil {
		mediaItemIds = nil
	}
	return
}

func (m *mediaItems) GetAll() ([]MediaItem, error) {
	query := `SELECT uuid, remote_id, base_url, mime_type, filename, description, downloaded,
					 local_path, local_filename, file_size, created_at, modified_at, synced_at, last_error
			  FROM media_items`

	var mediaItems []MediaItem
	mapper := func(row Scanner) (err error) {
		var mediaItem MediaItem
		err = mediaItemRowMapper(&mediaItem)(row)
		if err == nil {
			mediaItems = append(mediaItems, mediaItem)
		}
		return
	}

	err := m.sqlFuncs.Query(mapper, query)
	if err != nil {
		return nil, err
	}

	return mediaItems, nil
}

func (m *mediaItems) Truncate() error {
	return m.sqlFuncs.Truncate("media_items")
}

func (m *mediaItems) UpdateBaseUrl(remoteId string, baseUrl string) error {
	updateSql := "UPDATE media_items SET base_url = ? WHERE remote_id = ?;"
	return m.sqlFuncs.Exec(updateSql, baseUrl, remoteId)
}

func parseTime(dateTime sql.NullString, mediaTime *time.Time) (err error) {
	if !dateTime.Valid {
		return
	}

	*mediaTime, err = time.Parse(time.RFC3339Nano, dateTime.String)
	if err != nil {
		return
	}
	return
}

func mediaItemRowMapper(mediaItem *MediaItem) MapperFunc {
	return func(row Scanner) (err error) {
		var downloaded int
		var createdAt sql.NullString
		var modifiedAt sql.NullString
		var syncedAt sql.NullString
		var tempItem MediaItem

		err = row.Scan(
			&tempItem.Uuid,
			&tempItem.RemoteId,
			&tempItem.BaseUrl,
			&tempItem.MimeType,
			&tempItem.Filename,
			&tempItem.Description,
			&downloaded,
			&tempItem.LocalPath,
			&tempItem.LocalFilename,
			&tempItem.FileSize,
			&createdAt,
			&modifiedAt,
			&syncedAt,
			&tempItem.LastError,
		)
		if err != nil {
			return
		}

		tempItem.Downloaded = downloaded != 0

		err = parseTime(createdAt, &tempItem.CreatedAt)
		if err == nil {
			err = parseTime(modifiedAt, &tempItem.ModifiedAt)
			if err == nil {
				err = parseTime(syncedAt, &tempItem.SyncedAt)
			}
		}

		if err != nil {
			return
		}
		*mediaItem = tempItem

		return
	}
}
