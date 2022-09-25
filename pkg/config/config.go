package config

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
	oauthGoogle "golang.org/x/oauth2/google"
	oauthSpotify "golang.org/x/oauth2/spotify"
)

type Config struct {
	Domain         string
	Templates      string
	OauthProviders OauthProviders
}

func GetConfig(filePath string) (Config, error) {
	var cfg Config

	body, err := os.ReadFile(filePath)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(body, &cfg)
	if err != nil {
		return cfg, err
	}

	cfg.setDefaults()

	err = json.Unmarshal(body, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

type OauthProviders struct {
	Google  oauth2.Config
	Spotify oauth2.Config
}

func (c *Config) setDefaults() {
	c.Templates = "templates/*"

	c.OauthProviders.Google.Endpoint = oauthGoogle.Endpoint
	c.OauthProviders.Google.Scopes = []string{"https://www.googleapis.com/auth/userinfo.email"}
	c.OauthProviders.Google.RedirectURL = c.Domain + "/auth/google"

	c.OauthProviders.Spotify.Endpoint = oauthSpotify.Endpoint
	c.OauthProviders.Spotify.Scopes = []string{"user-library-read"}
	c.OauthProviders.Spotify.RedirectURL = c.Domain + "/auth/spotify"
}
