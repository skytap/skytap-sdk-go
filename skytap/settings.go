package skytap

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

const (
	DefaultBaseURL   = "https://cloud.skytap.com/"
	DefaultUserAgent = "skytap-sdk-go/" + version
)

type Settings struct {
	BaseUrl   string
	UserAgent string

	Credentials CredentialsProvider
}

func (s *Settings) Validate() error {
	var err *multierror.Error

	if s.BaseUrl == "" {
		err = multierror.Append(err, fmt.Errorf("the base URL must be provided"))
	}
	if s.UserAgent == "" {
		err = multierror.Append(err, fmt.Errorf("the user agent must be provided"))
	}
	if s.Credentials == nil {
		err = multierror.Append(err, fmt.Errorf("the credential provider must be provided"))
	}

	return err.ErrorOrNil()
}

func NewDefaultSettings(clientSettings ...ClientSetting) Settings {
	settings := Settings{
		BaseUrl:     DefaultBaseURL,
		UserAgent:   DefaultUserAgent,
		Credentials: NewNoOpCredentials(),
	}

	// Apply any custom settings
	for _, c := range clientSettings {
		c.Apply(&settings)
	}

	return settings
}

type ClientSetting interface {
	Apply(*Settings)
}

type withBaseUrl string
type withUserAgent string
type withCredentialsProvider struct{ cp CredentialsProvider }

func (w withBaseUrl) Apply(s *Settings) {
	s.BaseUrl = string(w)
}

func (w withUserAgent) Apply(s *Settings) {
	s.UserAgent = string(w)
}

func (w withCredentialsProvider) Apply(s *Settings) {
	s.Credentials = w.cp
}

func WithBaseUrl(BaseUrl string) ClientSetting {
	return withBaseUrl(BaseUrl)
}

func WithUserAgent(BaseUrl string) ClientSetting {
	return withUserAgent(BaseUrl)
}

func WithCredentialsProvider(credentialsProvider CredentialsProvider) ClientSetting {
	return withCredentialsProvider{credentialsProvider}
}
