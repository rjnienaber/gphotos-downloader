package database

import (
	"database/sql"
	"path"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

const GooglePhotosDatabaseFile = "google_photos.sqlite3"

// SQLite error codes
const ConstraintPrimaryError = 19
const ConstraintUniqueExtendedError = 2067

type PhotoDatabase struct {
	connection   *sql.DB
	databasePath string
	Settings     settings
	MediaItems   mediaItems
	Logger       utils.Logger
}

type Option func(svc *PhotoDatabase) error

func NewDatabase(opts ...Option) (db PhotoDatabase, err error) {
	db = PhotoDatabase{Logger: utils.NewLogger(utils.Info)}
	for _, opt := range opts {
		err = opt(&db)
		if err != nil {
			return
		}
	}

	if db.connection == nil {
		err = WithInMemoryConnection()(&db)
		if err != nil {
			return
		}
	}

	sqlFuncs := SqlFuncs{
		connection: db.connection,
		logger:     db.Logger,
	}

	db.Settings = settings{sqlFuncs: sqlFuncs, logger: db.Logger}
	db.MediaItems = mediaItems{sqlFuncs: sqlFuncs, logger: db.Logger}
	err = initialize(&db)
	return
}

func initialize(db *PhotoDatabase) (err error) {
	version, err := db.Settings.Version()
	if err != nil {
		sqliteError, ok := err.(sqlite3.Error)
		if ok && strings.Contains(sqliteError.Error(), "no such table") {
			db.Logger.Debug.Print("no version table found, initializing database")
			version = 0
		} else {
			return
		}
	}

	appDatabaseVersion := AppDatabaseVersion()
	if version != appDatabaseVersion {
		db.Logger.Info.Printf("db version: %d, expected %d. running migrations...\n", version, appDatabaseVersion)
		err = RunMigrations(db, version)
	} else {
		db.Logger.Debug.Print("database version in-sync with application")
	}
	return
}

func (db *PhotoDatabase) Close() error {
	return db.connection.Close()
}

func WithFileConnection(rootDir string, logger utils.Logger) Option {
	return func(db *PhotoDatabase) (err error) {
		db.databasePath = path.Join(rootDir, GooglePhotosDatabaseFile)
		logger.Debug.Printf("opening sqlite3 database at %s", db.databasePath)
		conn, err := sql.Open("sqlite3", "file:"+db.databasePath)
		conn.SetMaxOpenConns(1)
		if err != nil {
			return
		}
		logger.Trace.Printf("pinging open sqlite3 database at %s", db.databasePath)
		err = conn.Ping()
		if err != nil {
			return
		}
		db.connection = conn
		return
	}
}

func WithInMemoryConnection() Option {
	return func(db *PhotoDatabase) (err error) {
		conn, err := sql.Open("sqlite3", "file::memory:")
		if err == nil {
			db.connection = conn
		}
		return
	}
}

func WithLogger(logger utils.Logger) Option {
	return func(repos *PhotoDatabase) (err error) {
		repos.Logger = logger
		return
	}
}
