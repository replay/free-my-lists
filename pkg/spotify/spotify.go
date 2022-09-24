package spotify

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2"
	oauthSpotify "golang.org/x/oauth2/spotify"
)

type spotify struct {
	oauthCfg oauth2.Config
}

func newSpotify() spotify {
	return spotify{
		oauthCfg: oauth2.Config{
			ClientID:     "",
			ClientSecret: "",
			Scopes:       []string{"user-library-read"},
			Endpoint:     oauthSpotify.Endpoint,
			RedirectURL:  "https://free-my-lists.click/callback",
		},
	}
}

func (s *spotify) getMyTracks(ctx context.Context) {
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := s.oauthCfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := s.oauthCfg.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := s.oauthCfg.Client(ctx, tok)
	resp, err := client.Get("https://api.spotify.com/v1/me/tracks")
	if err != nil {
		panic(err)
	}

	tracks, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("tracks:\n%+v\n", tracks)
}
