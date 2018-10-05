package options

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
)

type DialSettings struct {
	User     string
	APIToken string
	Scheme   string
	Host     string
}

func (d *DialSettings) Validate() error {
	var err *multierror.Error

	if d.Scheme == "" {
		err = multierror.Append(err, fmt.Errorf("http scheme is missing"))
	}
	if d.Host == "" {
		err = multierror.Append(err, fmt.Errorf("http host is missing"))
	}
	if d.User == "" {
		err = multierror.Append(err, fmt.Errorf("http user is missing"))
	}
	if d.APIToken == "" {
		err = multierror.Append(err, fmt.Errorf("http APIToken is missing"))
	}
	return err.ErrorOrNil()
}

type ClientOption interface {
	Apply(*DialSettings)
}

type withUser string
type withApiToken string
type withScheme string
type withHost string

func (w withUser) Apply(o *DialSettings) {
	o.User = string(w)
}

func (w withApiToken) Apply(o *DialSettings) {
	o.APIToken = string(w)
}

func (w withScheme) Apply(o *DialSettings) {
	o.Scheme = string(w)
}

func (w withHost) Apply(o *DialSettings) {
	o.Host = string(w)
}

func WithUser(user string) ClientOption {
	return withUser(user)
}

func WithAPIToken(apiToken string) ClientOption {
	return withApiToken(apiToken)
}

func WithScheme(scheme string) ClientOption {
	return withScheme(scheme)
}

func WithHost(host string) ClientOption {
	return withHost(host)
}
