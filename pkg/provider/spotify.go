package provider

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

type spotifyProvider struct {
	provider
	spotify *spotify.Client
}

func newSpotify(p provider) spotifyProvider {
	return spotifyProvider{
		provider: p,
		spotify:  spotify.New(p.Client),
	}
}

func (s spotifyProvider) UserInfo(ctx context.Context) (UserInfo, error) {
	var resp UserInfo

	user, err := s.spotify.CurrentUser(ctx)
	if err != nil {
		return resp, err
	}

	resp.ID = user.ID
	resp.Name = user.DisplayName
	resp.Email = user.Email

	return resp, nil
}

func (s spotifyProvider) Lists(ctx context.Context) (Lists, error) {
	var resp Lists

	playlists, err := s.spotify.CurrentUsersPlaylists(ctx)
	if err != nil {
		return resp, err
	}

	for _, p := range playlists.Playlists {
		resp.lists = append(resp.lists, List{
			ID:   p.ID.String(),
			Name: p.Name,
		})
	}

	return resp, nil
}
