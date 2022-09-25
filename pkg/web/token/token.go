package token

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

type Type string

const (
	Google  Type = "google"
	Spotify Type = "spotify"
)

type Token struct {
	AccessToken *oauth2.Token `json:"access_token"`
	Provider    Type          `json:"provider"`
}

func NewToken(token *oauth2.Token, p Type) Token {
	return Token{
		AccessToken: token,
		Provider:    p,
	}
}

func (t Token) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

func Deserialize(data []byte) (Token, error) {
	var t Token
	err := json.Unmarshal(data, &t)
	if err != nil {
		return t, err
	}
	return t, nil
}
