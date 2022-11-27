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

	err := s.paginatedCall(func(token string) (string, error) {
		call := s.youtube.Playlists.List([]string{"snippet,contentDetails"})
		call = call.MaxResults(100)
		call = call.Mine(true)
		call = call.PageToken(token)
		response, err := call.Do()
		if err != nil {
			return "", err
		}

		for _, playlist := range response.Items {
			resp = append(resp, List{
				ID:   playlist.Id,
				Name: playlist.Snippet.Title,
			})
		}

		return response.NextPageToken, nil
	})

	return resp, err
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
		return resp, errors.New("playlist empty")
	}

	resp.Name = playlist.Items[0].Snippet.Title

	err = s.paginatedCall(func(token string) (string, error) {
		call := s.youtube.PlaylistItems.List([]string{"snippet,contentDetails"})
		call = call.PlaylistId(listID)
		call = call.MaxResults(100)
		call = call.PageToken(token)
		response, err := call.Do()
		if err != nil {
			return "", err
		}

		for _, item := range response.Items {
			resp.Tracks = append(resp.Tracks, Track{
				ID:   item.Snippet.ResourceId.VideoId,
				Name: item.Snippet.Title,
			})
		}

		return response.NextPageToken, nil
	})

	return resp, err
}

func (s googleProvider) paginatedCall(call func(string) (string, error)) error {
	nextPageToken, err := call("")
	if err != nil {
		return err
	}

	for nextPageToken != "" {
		nextPageToken, err = call(nextPageToken)
		if err != nil {
			return err
		}
	}

	return nil
}
