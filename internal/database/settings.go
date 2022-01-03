package database

import (
	"database/sql"
	"time"

	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

type settings struct {
	sqlFuncs SqlFuncs
	logger   utils.Logger
}

func (s *settings) Version() (version int, err error) {
	err = s.sqlFuncs.QueryValue("SELECT version FROM settings LIMIT 1", &version)
	return
}

func (s *settings) LastIndex() (lastIndex time.Time, err error) {
	var data sql.NullString
	err = s.sqlFuncs.QueryValue("SELECT last_index FROM settings LIMIT 1", &data)
	if err != nil {
		return
	}

	if data.Valid {
		lastIndex, err = time.Parse(time.RFC3339Nano, data.String)
	}

	return
}

func (s *settings) UpdateLastIndex(now time.Time) error {
	if now != (time.Time{}) {
		return s.sqlFuncs.Exec("UPDATE settings SET last_index = ?", now.Format(time.RFC3339Nano))
	} else {
		return s.sqlFuncs.Exec("UPDATE settings SET last_index = ?", nil)
	}
}

func (s *settings) Token() (token string, err error) {
	var data sql.NullString
	err = s.sqlFuncs.QueryValue("SELECT token FROM settings LIMIT 1", &data)
	if err != nil {
		return
	}

	if data.Valid {
		token = data.String
	}

	return
}

func (s *settings) UpdateToken(token string) (err error) {
	err = s.sqlFuncs.Exec("UPDATE settings SET token = ?", token)
	return
}
