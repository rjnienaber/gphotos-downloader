package database

import (
	"database/sql"
	"fmt"

	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

// Scanner is an interface used by Scan.
type Scanner interface {
	Scan(dest ...interface{}) error
}

type MapperFunc = func(row Scanner) error

type SqlFuncs struct {
	connection *sql.DB
	logger     utils.Logger
}

func (db *SqlFuncs) QueryValue(query string, args ...interface{}) (err error) {
	db.logger.Trace.Printf("query_row: %s", query)
	err = db.connection.QueryRow(query).Scan(args...)
	return
}

func (db *SqlFuncs) QueryRow(query string, args []interface{}, mapper MapperFunc) (err error) {
	db.logger.Trace.Printf("query_row: %s", query)
	row := db.connection.QueryRow(query, args...)
	err = row.Err()
	if err != nil {
		return
	}
	err = mapper(row)
	return
}

func (db *SqlFuncs) Exec(query string, args ...interface{}) (err error) {
	db.logger.Trace.Printf("exec query: '%s' %+v", query, args)
	_, err = db.connection.Exec(query, args...)
	return
}

func (db *SqlFuncs) Query(mapper MapperFunc, query string, args ...interface{}) (err error) {
	db.logger.Trace.Printf("exec query: '%s' %+v", query, args)
	rows, err := db.connection.Query(query, args...)
	if err != nil {
		return
	}
	defer utils.CheckClose(rows, &err)

	if rows.Err() != nil {
		err = rows.Err()
		return
	}

	for rows.Next() {
		err = mapper(rows)
		if err != nil {
			break
		}
	}
	return
}

func (db *SqlFuncs) Truncate(tableName string) (err error) {
	return db.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
}
