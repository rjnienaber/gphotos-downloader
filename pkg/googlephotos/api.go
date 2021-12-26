package googlephotos

import (
	"context"
	"github.com/rjnienaber/gphotos_downloader/pkg/googlephotos/models"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type PhotosApi struct {
	baseUrl string
	client  *http.Client
	logger  *log.Logger
}

type Options struct {
	BaseUrl               string
	Config                oauth2.Config
	Token                 *oauth2.Token
	Client                *http.Client
	TimeoutInMilliseconds int
	Logger                *log.Logger
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

func (api *PhotosApi) List(options models.PagingOptions) (mediaItems models.MediaItems, err error) {
	queryString := map[string][]string{}
	queryString["pageSize"] = []string{strconv.Itoa(options.Size)}
	queryString["pageToken"] = []string{options.Token}

	listUrl, err := api.buildUrl("/mediaItems", queryString)
	if err != nil {
		return
	}

	response, err := api.client.Get(listUrl.String())
	if err != nil {
		return
	}
	defer checkClose(response.Body, &err)

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
	response, err := api.client.Post(api.baseUrl+"/mediaItems:search", "application/json", bodyReader)
	if err != nil {
		return
	}
	defer checkClose(response.Body, &err)

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

func (api *PhotosApi) Download(item models.MediaItem) (filePath string, err error) {
	file, err := ioutil.TempFile("", "gphoto.*.tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer func(err *error) {
		if *err == nil {
			// download was successful, so we want to retain the file and just close it
			checkClose(file, err)
		} else {
			// download unsuccessful, so we want to clean up the temporary file
			// it's a temporary file so just log the error and return the original err
			removeError := os.Remove(file.Name())
			if removeError != nil {
				log.Printf("Error: %s\n", removeError.Error())
				return
			}
		}
	}(&err)

	downloadUrl := item.BaseUrl
	if item.IsPhoto() {
		downloadUrl += "=d"
	} else {
		downloadUrl += "=dv"
	}
	response, err := api.client.Get(downloadUrl)
	if err != nil {
		return
	}
	defer checkClose(response.Body, &err)

	_, err = file.ReadFrom(response.Body)
	if err != nil {
		return
	}

	filePath = file.Name()
	return
}

func (api *PhotosApi) buildUrl(resourceUrl string, queryString map[string][]string) (*url.URL, error) {
	fullUrl, err := url.Parse(api.baseUrl + resourceUrl)
	if err != nil {
		return nil, err
	}

	q := fullUrl.Query()
	for key, value := range queryString {
		for _, v := range value {
			if v != "" {
				q.Set(key, v)
			}
		}
	}
	fullUrl.RawQuery = q.Encode()
	return fullUrl, nil
}

func isSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func checkClose(c io.Closer, err *error) {
	cerr := c.Close()
	if *err == nil {
		*err = cerr
	}
}
