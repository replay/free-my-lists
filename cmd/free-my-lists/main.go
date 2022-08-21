package main

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

func main() {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     "82a836e2e06b4facb185aa520131a544",
		ClientSecret: "40f5df08b3c74c4fb9c84e11ad48108f",
		Scopes:       []string{"playlist-read-private"},
		Endpoint:     spotify.Endpoint,
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	resp, err := client.Get("https://api.spotify.com/v1/me/playlists")
	if err != nil {
		panic(err)
	}

	fmt.Printf("response: %+v\n", resp)
}
