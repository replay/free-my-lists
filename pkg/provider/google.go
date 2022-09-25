package provider

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

type googleProvider struct {
	provider
	google *people.Service
}

func newGoogle(ctx context.Context, p provider) (googleProvider, error) {
	g, err := people.NewService(ctx, option.WithHTTPClient(p.Client))
	if err != nil {
		return googleProvider{}, err
	}
	return googleProvider{p, g}, nil
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
