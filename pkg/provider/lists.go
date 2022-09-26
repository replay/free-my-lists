package provider

import (
	"context"
)

type ListsProvider interface {
	AuthProvider
	Lists(context.Context) (Lists, error)
	ListDetails(context.Context, string) (ListDetails, error)
}

type Lists []List

type List struct {
	ID   string
	Name string
}

type ListDetails struct {
	ID     string
	Name   string
	Tracks []Track
}

type Track struct {
	ID   string
	Name string
}
