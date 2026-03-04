package zammad

import (
	"fmt"
	"net/http"
)

type (
	// Client is used to query Zammad. It is safe to use concurrently. If you (inadvertly) added
	// multiple authencation options that will be applied in the order, basic auth, token based, and
	// then oauth. Where the last one set, wins.
	client[T any] struct {
		Client   Doer
		Username string
		Password string
		Token    string
		OAuth    string
		Url      string
		FromFunc func() string
	}
	Client = client[struct{}]

	// ErrorResponse is the response returned by Zammad when an error occured.
	ErrorResponse struct {
		Description      string `json:"error"`
		DescriptionHuman string `json:"error_human"`
	}

	// Doer is an interface that allows mimicking a *http.Client.
	Doer interface {
		Do(*http.Request) (*http.Response, error)
	}
)

func (r *ErrorResponse) Error() string {
	return fmt.Sprint(r.Description)
}
