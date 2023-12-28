package digipoauth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type Credentials struct {
	Username  string
	Password  string
	OTPSecret string
}

type Option interface {
	Apply(instance interface{}) error
}

type OptionFunc func(instance interface{}) error

func (f OptionFunc) Apply(instance interface{}) error {
	return f(instance)
}

// LoginMethod is the method to connect to digiposte.
type LoginMethod interface {
	Login(ctx context.Context, creds *Credentials) (*oauth2.Token, []*http.Cookie, error)
}

type LoginMethodFunc func(ctx context.Context, creds *Credentials) (*oauth2.Token, []*http.Cookie, error)

func (f LoginMethodFunc) Login(ctx context.Context, creds *Credentials) (*oauth2.Token, []*http.Cookie, error) {
	return f(ctx, creds)
}

type InvalidOptionError struct {
	Name string
	Err  error
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf("option %q: %v", e.Name, e.Err)
}

func (e *InvalidOptionError) Unwrap() error {
	return e.Err
}

func (e *InvalidOptionError) Apply(interface{}) error {
	return e
}
