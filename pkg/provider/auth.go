package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replay/free-my-lists/pkg/config"
	"github.com/replay/free-my-lists/pkg/web/token"
	"golang.org/x/oauth2"
)

type AuthProvider interface {
	UserInfo(context.Context) (UserInfo, error)
	ListsProvider() (ListsProvider, error)
}

type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type provider struct {
	Cfg *oauth2.Config
	*http.Client
	token token.Token
}

func NewAuthProvider(ctx context.Context, cfg config.Config, t token.Token) (AuthProvider, error) {
	pCfg := Config(cfg, t.Provider)
	p := provider{
		Cfg:    pCfg,
		Client: pCfg.Client(ctx, t.AccessToken),
		token:  t,
	}

	switch t.Provider {
	case token.Google:
		g, err := newGoogle(ctx, p)
		if err != nil {
			return nil, err
		}
		return g, nil
	case token.Spotify:
		return newSpotify(p), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", t.Provider)
	}
}

func (p provider) ListsProvider() (ListsProvider, error) {
	switch p.token.Provider {
	case token.Google:
		g, err := newGoogle(context.Background(), p)
		if err != nil {
			return nil, err
		}
		return g, nil
	case token.Spotify:
		return newSpotify(p), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", p.token.Provider)
	}
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
