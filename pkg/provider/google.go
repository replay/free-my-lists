package provider

import (
	"context"
	"errors"

	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
	"google.golang.org/api/youtube/v3"
)

type googleProvider struct {
	provider
	google  *people.Service
	youtube *youtube.Service
}

func newGoogle(ctx context.Context, p provider) (googleProvider, error) {
	g, err := people.NewService(ctx, option.WithHTTPClient(p.Client))
	if err != nil {
		return googleProvider{}, err
	}

	y, err := youtube.NewService(ctx, option.WithHTTPClient(p.Client))
	if err != nil {
		return googleProvider{}, err
	}

	return googleProvider{p, g, y}, nil
}

func (g googleProvider) UserInfo(_ context.Context) (UserInfo, error) {
	var resp UserInfo

	user, err := g.google.People.Get("people/me").PersonFields("names,emailAddresses").Do()
	if err != nil {
		return resp, err
	}

	if len(user.Names) > 0 {
		resp.Name = user.Names[0].DisplayName
	}
	if len(user.EmailAddresses) > 0 {
		resp.ID = user.EmailAddresses[0].Value
		resp.Email = resp.ID
	}

	return resp, nil
}

func (s googleProvider) Lists(ctx context.Context) (Lists, error) {
	resp := Lists{
		{
			ID:   "LM",
			Name: "Liked Music",
		},
	}

	response, err := s.youtube.Playlists.List([]string{"snippet,contentDetails"}).MaxResults(100).Mine(true).Do()
	if err != nil {
		return resp, err
	}

	for _, playlist := range response.Items {
		resp = append(resp, List{
			ID:   playlist.Id,
			Name: playlist.Snippet.Title,
		})
	}

	return resp, nil
}

func (s googleProvider) ListDetails(ctx context.Context, listID string) (ListDetails, error) {
	resp := ListDetails{
		ID: listID,
	}

	playlist, err := s.youtube.Playlists.List([]string{"snippet,contentDetails"}).Id(listID).Do()
	if err != nil {
		return resp, err
	}

	if len(playlist.Items) == 0 {
		return resp, errors.New("playlist not found")
	}

	resp.Name = playlist.Items[0].Snippet.Title

	response, err := s.youtube.PlaylistItems.List([]string{"snippet,contentDetails"}).PlaylistId(listID).MaxResults(100).Do()
	if err != nil {
		return resp, err
	}

	for _, item := range response.Items {
		resp.Tracks = append(resp.Tracks, Track{
			ID:   item.Snippet.ResourceId.VideoId,
			Name: item.Snippet.Title,
		})
	}

	return resp, nil
}
