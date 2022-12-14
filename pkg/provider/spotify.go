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

	err := s.paginatedCall(func(offset int) (int, bool, error) {
		limit := 50
		opts := []spotify.RequestOption{
			spotify.Limit(limit),
			spotify.Offset(offset),
		}

		lists, err := s.spotify.CurrentUsersPlaylists(ctx, opts...)
		if err != nil {
			return 0, false, err
		}

		for _, l := range lists.Playlists {
			resp = append(resp, List{
				ID:   l.ID.String(),
				Name: l.Name,
			})
		}

		return offset + limit, len(lists.Playlists) >= limit, nil
	})

	return resp, err
}

func (s spotifyProvider) ListDetails(ctx context.Context, listID string) (ListDetails, error) {
	var resp ListDetails

	list, err := s.spotify.GetPlaylist(ctx, spotify.ID(listID))
	if err != nil {
		return resp, err
	}

	for _, t := range list.Tracks.Tracks {
		resp.Tracks = append(resp.Tracks, Track{
			ID:   t.Track.ID.String(),
			Name: t.Track.Name,
		})
	}

	return resp, err
}

func (s spotifyProvider) paginatedCall(call func(int) (int, bool, error)) error {
	offset, next, err := call(0)
	if err != nil {
		return err
	}

	for next {
		offset, next, err = call(offset)
		if err != nil {
			return err
		}
	}

	return nil
}
