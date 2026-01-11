package oauth2

type GoogleClientSecret struct {
	Installed InstalledCredentials `json:"installed"`
}

type InstalledCredentials struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	AuthUri      string   `json:"auth_uri"`
	TokenUri     string   `json:"token_uri"`
	RedirectUris []string `json:"redirect_uris"`
}
