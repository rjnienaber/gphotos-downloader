package database

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func AppDatabaseVersion() int {
	dir, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range dir {
		if strings.HasSuffix(entry.Name(), ".sql") {
			count++
		}
	}

	return count
}

func RunMigrations(db *PhotoDatabase, currentVersion int) (err error) {
	for i := currentVersion; i < AppDatabaseVersion(); i++ {
		migrationFilename := fmt.Sprintf("migrations/version_%03d.sql", i)
		db.Logger.Debug.Printf("retrieving migration file '%s'", migrationFilename)
		var sqlStatements []byte
		sqlStatements, err = migrationFiles.ReadFile(migrationFilename)
		if err != nil {
			break
		}

		nextVersion := i + 1
		db.Logger.Info.Printf("migrating from version %d to version %d", i, nextVersion)
		_, err = db.connection.Exec(string(sqlStatements))
		if err != nil {
			break
		}
		updateVersionSql := fmt.Sprintf("UPDATE settings SET version = %d", nextVersion)
		_, err = db.connection.Exec(updateVersionSql)
		if err != nil {
			break
		}

	}
	return
}
