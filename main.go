package main

import (
	"net"
	"os"

	"github.com/rjnienaber/gphotos_downloader/internal/database"
	photoOauth "github.com/rjnienaber/gphotos_downloader/internal/oauth2"
	"github.com/rjnienaber/gphotos_downloader/internal/services"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

func wireUp(clientSecretPath string, dbPath string, db database.PhotoDatabase, logger utils.Logger) (services.DownloadService, services.UndownloadedService, services.SyncService) {
	tokenService, err := photoOauth.NewTokenService(clientSecretPath, db, logger)
	if err != nil {
		logger.Error.Fatal(err)
	}

	token, err := tokenService.LoadToken()
	if err != nil {
		logger.Error.Fatal(err)
	}

	photosApi := googlephotos.NewPhotosApi(googlephotos.Options{
		Config: tokenService.Config,
		Token:  token,
		Logger: logger,
	})

	retryFactory := services.NewExponentialRetryFactory(net.OpError{}, new(net.OpError), net.DNSError{}, new(net.DNSError))
	downloader := services.NewDownloadService(&photosApi, db, dbPath,
		services.WithLogger(logger),
		services.WithRetryFactory(retryFactory),
		services.WithMaxWorkers(5),
	)

	undownloadedService := services.NewUndownloadedService(&photosApi, db, &downloader, logger)
	syncService := services.NewSyncService(&photosApi, db, &downloader, logger)

	return downloader, undownloadedService, syncService
}

func main() {
	clientSecretPath := os.Args[1]
	dbPath := os.Args[2]

	logger := utils.NewLogger(utils.Debug)
	db, err := database.NewDatabase(
		database.WithFileConnection(dbPath, logger),
		database.WithLogger(logger),
	)
	if err != nil {
		logger.Error.Fatal(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			logger.Error.Print(err)
		}
	}()

	downloader, undownloadedService, syncService := wireUp(clientSecretPath, dbPath, db, logger)

	err = undownloadedService.Update()
	if err != nil {
		downloader.Finish()
		logger.Error.Fatal(err)
		return
	}

	err = syncService.Sync()
	if err != nil {
		downloader.Finish()
		logger.Error.Fatal(err)
		return
	}

	downloader.Finish()

	logger.Info.Print("sync completed")
}
