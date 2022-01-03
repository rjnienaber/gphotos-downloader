package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	db := CreateTestDatabase(t)

	version, err := db.Settings.Version()
	assert.NoError(t, err)
	assert.Equal(t, AppDatabaseVersion(), version)
}

func TestLastIndex(t *testing.T) {
	db := CreateTestDatabase(t)

	lastIndex, err := db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.Equal(t, time.Time{}.String(), lastIndex.String())

	now := time.Now()
	err = db.Settings.UpdateLastIndex(now)
	assert.NoError(t, err)

	lastIndex, err = db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), lastIndex.UnixMilli())
}

func TestLastIndexUpdateWithNil(t *testing.T) {
	db := CreateTestDatabase(t)

	lastIndex, err := db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.Equal(t, time.Time{}.String(), lastIndex.String())

	err = db.Settings.UpdateLastIndex(time.Time{})
	assert.NoError(t, err)

	lastIndex, err = db.Settings.LastIndex()
	assert.NoError(t, err)
	assert.Equal(t, time.Time{}, lastIndex)

}

func TestToken(t *testing.T) {
	db := CreateTestDatabase(t)

	token, err := db.Settings.Token()
	assert.NoError(t, err)
	assert.Equal(t, "", token)

	err = db.Settings.UpdateToken("abc123")
	assert.NoError(t, err)

	token, err = db.Settings.Token()
	assert.NoError(t, err)
	assert.Equal(t, "abc123", token)

}
