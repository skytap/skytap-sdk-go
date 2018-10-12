package skytap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultSettings(t *testing.T) {
	settings := NewDefaultSettings()

	assert.Equal(t, DefaultBaseURL, settings.BaseUrl)
	assert.Equal(t, DefaultUserAgent, settings.UserAgent)

	if assert.NotNil(t, settings.Credentials) {
		assert.IsType(t, &NoOpCredentials{}, settings.Credentials)
	}
}

func TestNewDefaultSettingsWithOpts(t *testing.T) {
	baseUrl := "https://url.com"
	userAgent := "testclient/1.0.0"
	username := "user"
	password := "password"

	settings := NewDefaultSettings(
		WithBaseUrl(baseUrl),
		WithCredentialsProvider(NewPasswordCredentials(username, password)))

	assert.Equal(t, baseUrl, settings.BaseUrl)
	assert.Equal(t, DefaultUserAgent, settings.UserAgent)

	if assert.NotNil(t, settings.Credentials) {
		assert.IsType(t, &PasswordCredentials{}, settings.Credentials)
	}

	settings = NewDefaultSettings(WithUserAgent(userAgent))

	assert.Equal(t, DefaultBaseURL, settings.BaseUrl)
	assert.Equal(t, userAgent, settings.UserAgent)

	if assert.NotNil(t, settings.Credentials) {
		assert.IsType(t, &NoOpCredentials{}, settings.Credentials)
	}
}
