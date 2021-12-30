package oauth2

import (
	"bufio"
	"context"
	"crypto/rand"
	json2 "encoding/json"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/rjnienaber/gphotos_downloader/internal/database"
	"github.com/rjnienaber/gphotos_downloader/pkg/utils"
	"golang.org/x/oauth2"
)

type TokenService struct {
	Config oauth2.Config
	db     database.PhotoDatabase
	logger utils.Logger
}

func NewTokenService(configFilePath string, db database.PhotoDatabase, logger utils.Logger) (service TokenService, err error) {
	config, err := loadConfig(configFilePath, logger)
	if err == nil {
		service = TokenService{Config: config, db: db, logger: logger}
	}
	return
}

func loadClientSecret(filePath string, logger utils.Logger) (clientSecret GoogleClientSecret, err error) {
	logger.Debug.Print("loading google client secret file")
	json, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	err = json2.Unmarshal(json, &clientSecret)
	return
}

func loadConfig(filePath string, logger utils.Logger) (config oauth2.Config, err error) {
	clientSecret, err := loadClientSecret(filePath, logger)
	if err != nil {
		return
	}

	logger.Debug.Print("creating oauth2 config")
	installed := clientSecret.Installed
	return oauth2.Config{
		ClientID:     installed.ClientID,
		ClientSecret: installed.ClientSecret,
		RedirectURL:  installed.RedirectUris[0],
		Endpoint: oauth2.Endpoint{
			AuthURL:   installed.AuthUri,
			TokenURL:  installed.TokenUri,
			AuthStyle: 0,
		},
		Scopes: []string{
			"https://www.googleapis.com/auth/photoslibrary.readonly",
			"https://www.googleapis.com/auth/photoslibrary.sharing",
		},
	}, nil
}

func (ts *TokenService) LoadToken() (token *oauth2.Token, err error) {
	ts.logger.Debug.Print("loading token from database")
	dbToken, err := ts.db.Settings.Token()
	if err != nil {
		return
	}

	if dbToken == "" {
		ts.logger.Debug.Print("no token in database, asking to authorize")
		token, err = ts.authorizeToken()
	} else {
		ts.logger.Trace.Print("marshalling database json token to oauth token")
		err = json2.Unmarshal([]byte(dbToken), &token)
	}

	return
}

func (ts *TokenService) saveToken(token *oauth2.Token) (err error) {
	tokenBytes, err := json2.Marshal(token)
	if err != nil {
		return
	}
	err = ts.db.Settings.UpdateToken(string(tokenBytes))
	return
}

func (ts *TokenService) authorizeToken() (token *oauth2.Token, err error) {
	fullUrl, err := buildAuthorizationUrl(ts.Config)
	if err != nil {
		return
	}

	ts.logger.Default.Printf("Please go here and authorize, %s\n", fullUrl)
	ts.logger.Default.Printf("Paste the response token here: ")
	reader := bufio.NewReader(os.Stdin)
	authCode, _ := reader.ReadString('\n')

	ts.logger.Debug.Print("using authcode to request new token json")
	token, err = ts.Config.Exchange(context.TODO(), authCode)
	if err == nil {
		err = ts.saveToken(token)
	}

	return
}

func buildAuthorizationUrl(config oauth2.Config) (authUrl string, err error) {
	state, err := generateState()
	if err != nil {
		return
	}

	authUrl = config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
	return
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateState() (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	length := 30
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
