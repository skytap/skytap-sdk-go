package skytap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/opencredo/skytap-sdk-go-internal/options"
	"io"
	"net/http"
	"net/url"
)

const userAgent = "skytap-sdk-go/0.1"

// Client is a client to manage and configure the skytap cloud
//
type Client struct {
	hc       *http.Client
	baseUrl  *url.URL
	settings *options.DialSettings
}

// NewClient creates a Skytab cloud client
//
func NewClient(ctx context.Context, opts ...options.ClientOption) (*Client, error) {
	// Transport configuration
	transport := http.DefaultTransport
	settings, err := newSettings(opts)
	if err != nil {
		return nil, err
	}

	return &Client{
		hc: &http.Client{
			Transport: transport,
		},
		baseUrl: &url.URL{
			Scheme: settings.Scheme,
			Host:   settings.Host,
		},
		settings: settings,
	}, nil
}

// Close a existing client
//
// There is not requirement of closing a client when a program exit.
func (c *Client) Close() error {
	c.hc = nil
	return nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.baseUrl.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.SetBasicAuth(c.settings.User, c.settings.APIToken)
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	//TODO handle default error codes, resource wait and authentication issues.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = errors.New(resp.Status)
	}

	return resp, err
}

func newSettings(opts []options.ClientOption) (*options.DialSettings, error) {
	var o options.DialSettings
	for _, opt := range opts {
		opt.Apply(&o)
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}
	return &o, nil
}
