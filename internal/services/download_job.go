package services

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
)

type DownloadJob struct {
	Id           string
	api          googlephotos.Downloader
	db           database.PhotoDatabase
	retryFactory RetryFactory
	logger       utils.Logger
	rootDir      string
}

func (j *DownloadJob) Process() {
	err := j.process()

	if err != nil {
		j.logger.Error.Printf("(id: %s) saving last error", j.Id)
		err = j.db.MediaItems.MarkAsErrored(j.Id, err)
		if err != nil {
			j.logger.Error.Printf("(id: %s) saving last error failed: %s", j.Id, err.Error())
		}
	}
}

func (j *DownloadJob) process() (err error) {
	j.logger.Trace.Printf("(id: %s) retrieving db item", j.Id)
	item, err := j.db.MediaItems.Get(j.Id)
	if err != nil {
		j.logger.Error.Printf("(id: %s) database retrieval failed: %s", j.Id, err.Error())
		return
	}

	if item.Downloaded {
		return
	}

	retry := j.retryFactory.Create()
	var tmpFilepath string
	for {
		tmpFilepath, err = j.downloadItem(item)
		if err != nil {
			if retry.ShouldRetry(err) {
				retry.Wait()
				continue
			}
			return
		}
		break
	}

	relativePath := filepath.Join(item.LocalPath, item.LocalFilename)
	itemFilepath := filepath.Join(j.rootDir, relativePath)

	j.logger.Trace.Printf("(id: %s) getting stat of root dir '%s'", j.Id, j.rootDir)
	stat, err := os.Stat(j.rootDir)
	if err != nil {
		j.logger.Error.Printf("(id: %s) getting stat of root dir '%s' failed: %s", j.Id, j.rootDir, err.Error())
		return
	}

	dir := filepath.Dir(itemFilepath)
	j.logger.Trace.Printf("(id: %s) ensuring item directory path for '%s'", j.Id, dir)
	err = os.MkdirAll(dir, stat.Mode())
	if err != nil {
		j.logger.Error.Printf("(id: %s) ensuring item directory path for '%s' failed: %s", j.Id, dir, err.Error())
		return
	}

	j.logger.Debug.Printf("(id: %s) moving file from '%s' to '%s'", j.Id, tmpFilepath, relativePath)
	// will overwrite an existing file
	err = os.Rename(tmpFilepath, itemFilepath)
	if err != nil {
		j.logger.Trace.Printf("(id: %s) moving failed, copying file from '%s' to '%s'", j.Id, tmpFilepath, relativePath)
		err = j.copyFile(tmpFilepath, itemFilepath)
		if err != nil {
			return
		}

		j.logger.Trace.Printf("(id: %s) deleting temporary file '%s'", j.Id, tmpFilepath)
		err = os.Remove(tmpFilepath)
		if err != nil {
			j.logger.Debug.Printf("(id: %s) deleting temporary file '%s' failed: %s", j.Id, tmpFilepath, err.Error())
			return
		}
	}

	j.logger.Trace.Printf("(id: %s) getting file size for '%s'", j.Id, relativePath)
	fileStat, err := os.Stat(itemFilepath)
	if err != nil {
		j.logger.Error.Printf("(id: %s) getting file size for '%s' failed: %s", j.Id, relativePath, err.Error())
		return
	}

	j.logger.Debug.Printf("(id: %s) marking file '%s' as synced", j.Id, relativePath)
	err = j.db.MediaItems.MarkAsSynced(j.Id, fileStat.Size())
	if err != nil {
		j.logger.Error.Printf("(id: %s) marking file '%s' as synced failed: %s", j.Id, relativePath, err.Error())
		return
	}

	j.logger.Info.Printf("(id: %s) downloaded '%s'", j.Id, relativePath)
	return
}

func (j *DownloadJob) downloadItem(item database.MediaItem) (string, error) {
	j.logger.Debug.Printf("(id: %s) downloading content of remote id '%s'", j.Id, item.RemoteId)
	tmpFilepath, downloadError := j.api.Download(item.BaseUrl, item.IsPhoto())
	if downloadError == nil {
		return tmpFilepath, nil
	}

	j.logger.Error.Printf("(id: %s) downloading content of remote id '%s' failed: %s", j.Id, item.RemoteId, downloadError.Error())

	// check for 403, which likely means the BaseUrl has changed
	apiError, ok := downloadError.(models.ApiError)
	if !ok || apiError.StatusCode != 403 {
		return "", downloadError
	}

	j.logger.Debug.Printf("(id: %s) getting new base url", j.Id)
	apiItem, err := j.api.Get(item.RemoteId)
	if err != nil {
		j.logger.Error.Printf("(id: %s) getting new base url failed: %s", j.Id, err.Error())
		return "", err
	}

	if item.BaseUrl == apiItem.BaseUrl {
		j.logger.Debug.Printf("(id: %s) fresh base url is the same as old one, nothing to do here", j.Id)
		return "", downloadError
	}

	j.logger.Debug.Printf("(id: %s) updating base url in database", j.Id)
	err = j.db.MediaItems.UpdateBaseUrl(item.RemoteId, apiItem.BaseUrl)
	if err != nil {
		j.logger.Debug.Printf("(id: %s) updating base url in database failed: %s", j.Id, err.Error())
		return "", err
	}

	j.logger.Debug.Printf("(id: %s) attempting content download of remote id '%s' with new base url", j.Id, item.RemoteId)
	tmpFilepath, err = j.api.Download(apiItem.BaseUrl, item.IsPhoto())
	if err != nil {
		j.logger.Error.Printf("(id: %s) downloading content of remote id '%s' failed: %s", j.Id, item.RemoteId, err.Error())
		return "", err
	}
	return tmpFilepath, nil
}

func (j *DownloadJob) copyFile(src, dest string) (err error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		j.logger.Debug.Printf("(id: %s) opening source file '%s' failed: %s", j.Id, src, err.Error())
		return
	}
	defer utils.CheckClose(sourceFile, &err)

	newFile, err := os.Create(dest)
	if err != nil {
		j.logger.Debug.Printf("(id: %s) creating destination file '%s' failed: %s", j.Id, dest, err.Error())
		return
	}
	defer utils.CheckClose(newFile, &err)

	_, err = io.Copy(newFile, sourceFile)
	if err != nil {
		j.logger.Debug.Printf("(id: %s) copying from '%s' to '%s' failed: %s", j.Id, src, dest, err.Error())
		return
	}
	return
}
