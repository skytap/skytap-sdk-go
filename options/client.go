package options

type DialSettings struct {
	Endpoint string
	User     string
	APIToken string
}

func (*DialSettings) Validate() error {
	return nil
}

type ClientOption interface {
	Apply(*DialSettings)
}

type withUser string
type withApiToken string

func (w withUser) Apply(o *DialSettings) {
	o.User = string(w)
}

func (w withApiToken) Apply(o *DialSettings) {
	o.APIToken = string(w)
}

func WithUser(user string) ClientOption {
	return withUser(user)
}

func WithAPIToken(apiToken string) ClientOption {
	return withApiToken(apiToken)
}
