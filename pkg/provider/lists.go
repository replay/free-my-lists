package provider

import (
	"context"
)

type ListsProvider interface {
	AuthProvider
	Lists(context.Context) (Lists, error)
}

type Lists struct {
	lists []List
}

type List struct {
	ID   string
	Name string
}
