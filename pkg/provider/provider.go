package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replay/free-my-lists/pkg/config"
	"github.com/replay/free-my-lists/pkg/web/token"
	"golang.org/x/oauth2"
)

var userInfoUrl = map[token.Type]string{
	token.Google:  "https://www.googleapis.com/oauth2/v3/userinfo",
	token.Spotify: "https://api.spotify.com/v1/me",
}

type ProviderClient struct {
	Cfg *oauth2.Config
	*http.Client
	token token.Token
}

func NewClient(ctx context.Context, cfg config.Config, t token.Token) ProviderClient {
	pCfg := Config(cfg, t.Provider)
	return ProviderClient{
		Cfg:    pCfg,
		Client: pCfg.Client(ctx, t.AccessToken),
		token:  t,
	}
}

func (p ProviderClient) UserInfo() (*http.Response, error) {
	url, ok := userInfoUrl[p.token.Provider]
	if !ok {
		return nil, fmt.Errorf("Unknown provider type: %s", p.token.Provider)
	}

	return p.Client.Get(url)
}

func Config(cfg config.Config, t token.Type) *oauth2.Config {
	switch t {
	case token.Google:
		return &cfg.OauthProviders.Google
	case token.Spotify:
		return &cfg.OauthProviders.Spotify
	}
	return nil
}
