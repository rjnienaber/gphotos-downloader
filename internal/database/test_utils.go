package database

import (
	"fmt"
	"testing"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func CreateTestDatabase(t *testing.T) PhotoDatabase {
	db, err := NewDatabase(WithLogger(utils.NewLogger(utils.Silent)))
	assert.NoError(t, err)
	return db
}

func CreateTestMediaItem(t *testing.T) MediaItem {
	remoteId, err := uuid.NewV4()
	assert.NoError(t, err)

	localFileUuid, err := uuid.NewV4()
	assert.NoError(t, err)
	fileName := fmt.Sprintf("DSC_01234_%s.png", localFileUuid.String())

	return MediaItem{
		RemoteId:      remoteId.String(),
		BaseUrl:       "https://home.com",
		MimeType:      "image/png",
		Filename:      fileName,
		Description:   "my description",
		Downloaded:    true,
		LocalPath:     "2012/12/12",
		LocalFilename: fileName,
		FileSize:      2345,
		CreatedAt:     timeMustParse(t, "2012-12-12T19:54:05Z"),
		ModifiedAt:    timeMustParse(t, "2012-12-12T06:54:05Z"),
		SyncedAt:      timeMustParse(t, "2012-12-03T19:54:05Z"),
	}
}

func timeMustParse(t *testing.T, dateTime string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, dateTime)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsedTime)
	return parsedTime
}
