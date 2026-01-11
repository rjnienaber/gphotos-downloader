package googlephotos

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
	"golang.org/x/oauth2"
)

type PhotosApi struct {
	baseUrl string
	client  *http.Client
	logger  utils.Logger
}

type Options struct {
	BaseUrl               string
	Config                oauth2.Config
	Token                 *oauth2.Token
	Client                *http.Client
	TimeoutInMilliseconds int
	Logger                utils.Logger
}

func NewPhotosApi(options Options) PhotosApi {
	client := options.Client
	if client == nil {
		client = options.Config.Client(context.Background(), options.Token)
	}

	if options.TimeoutInMilliseconds > 0 {
		client.Timeout = time.Duration(options.TimeoutInMilliseconds) * time.Millisecond
	}

	baseUrl := options.BaseUrl
	if baseUrl == "" {
		baseUrl = "https://photoslibrary.googleapis.com/v1"
	}

	return PhotosApi{
		baseUrl: baseUrl,
		client:  client,
		logger:  options.Logger,
	}
}

func (api *PhotosApi) Get(mediaItemId string) (mediaItem models.MediaItem, err error) {
	getUrl, err := api.buildUrl(fmt.Sprintf("/mediaItems/%s", mediaItemId), map[string][]string{})
	if err != nil {
		return
	}

	api.logger.Debug.Printf("getting media item from %s\n", getUrl.String())
	response, err := api.client.Get(getUrl.String())
	if err != nil {
		return
	}
	defer utils.CheckClose(response.Body, &err)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	if isSuccessResponse(response) {
		mediaItem, err = models.DeserializeMediaItemJson(responseBody)
		return
	}

	return models.MediaItem{}, models.ParseErrorReponse(response, responseBody)
}

func (api *PhotosApi) BatchGet(mediaItemIds []string) (mediaItems models.MediaItemsResult, err error) {
	queryString := map[string][]string{}
	queryString["mediaItemIds"] = mediaItemIds

	batchGetUrl, err := api.buildUrl("/mediaItems:batchGet", queryString)
	if err != nil {
		return
	}

	api.logger.Debug.Printf("getting list of media items from %s\n", batchGetUrl.String())
	response, err := api.client.Get(batchGetUrl.String())
	if err != nil {
		return
	}
	defer utils.CheckClose(response.Body, &err)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	if isSuccessResponse(response) {
		mediaItems, err = models.DeserializeMediaItemsResultJson(responseBody, mediaItemIds)
		return
	}

	return models.MediaItemsResult{}, models.ParseErrorReponse(response, responseBody)
}

func (api *PhotosApi) List(options models.PagingOptions) (mediaItems models.MediaItems, err error) {
	queryString := map[string][]string{}
	queryString["pageSize"] = []string{strconv.Itoa(options.Size)}
	queryString["pageToken"] = []string{options.Token}

	listUrl, err := api.buildUrl("/mediaItems", queryString)
	if err != nil {
		return
	}

	api.logger.Debug.Printf("getting list of media items from %s\n", listUrl.String())
	response, err := api.client.Get(listUrl.String())
	if err != nil {
		return
	}
	defer utils.CheckClose(response.Body, &err)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	if isSuccessResponse(response) {
		mediaItems, err = models.DeserializeMediaItemsJson(responseBody)
		return
	}

	return models.MediaItems{}, models.ParseErrorReponse(response, responseBody)
}

func (api *PhotosApi) Search(options models.SearchOptions) (mediaItems models.MediaItems, err error) {
	bodyReader, err := options.Serialize()
	if err != nil {
		return
	}

	searchUrl := api.baseUrl + "/mediaItems:search"
	api.logger.Debug.Printf("running search against %s\n", searchUrl)
	response, err := api.client.Post(searchUrl, "application/json", bodyReader)
	if err != nil {
		return
	}
	defer utils.CheckClose(response.Body, &err)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	if isSuccessResponse(response) {
		mediaItems, err = models.DeserializeMediaItemsJson(responseBody)
		return
	}

	return models.MediaItems{}, models.ParseErrorReponse(response, responseBody)
}

func (api *PhotosApi) Download(tmpDir string, baseUrl string, isPhoto bool) (filePath string, err error) {
	file, err := ioutil.TempFile(tmpDir, "gphoto.*.tmp")
	if err != nil {
		api.logger.Error.Print(err)
	}
	defer func(err *error) {
		if *err == nil {
			// download was successful, so we want to retain the file and just close it
			utils.CheckClose(file, err)
		} else {
			// download unsuccessful, so we want to clean up the temporary file
			removeError := os.Remove(file.Name())
			if removeError != nil {
				// it's a temporary file so just log the error and return the original err
				api.logger.Error.Printf("Error: %s\n", removeError.Error())
				return
			}
		}
	}(&err)

	if isPhoto {
		baseUrl += "=d"
	} else {
		baseUrl += "=dv"
	}

	api.logger.Trace.Printf("retrieving media item from %s\n", baseUrl)
	response, err := api.client.Get(baseUrl)
	if err != nil {
		return
	}
	defer utils.CheckClose(response.Body, &err)

	if !isSuccessResponse(response) {
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		return "", models.NewApiError(response, responseBody)
	}

	_, err = file.ReadFrom(response.Body)
	if err != nil {
		return
	}

	filePath = file.Name()
	return
}

func (api *PhotosApi) buildUrl(resourceUrl string, queryString map[string][]string) (fullUrl *url.URL, err error) {
	fullUrl, err = utils.BuildUrl(api.baseUrl+resourceUrl, queryString)
	return
}

func isSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
