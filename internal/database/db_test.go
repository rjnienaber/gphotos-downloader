package database

import (
	"io/fs"
	"os"
	"testing"

	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestWithFileConnection(t *testing.T) {
	logger := utils.NewLogger(utils.Silent)
	db, err := NewDatabase(
		WithLogger(logger),
		WithFileConnection(os.TempDir(), logger),
	)
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			_, ok := err.(*fs.PathError)
			if ok {
				return
			}
		}
		assert.NoError(t, err)
	}(db.databasePath)

	_, err = os.Stat(db.databasePath)
	assert.NoError(t, err)

	version, err := db.Settings.Version()
	assert.NoError(t, err)
	assert.Equal(t, AppDatabaseVersion(), version)
}
