package skytap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultSettings(t *testing.T) {
	settings := NewDefaultSettings()

	assert.Equal(t, DefaultBaseURL, settings.baseURL)
	assert.Equal(t, DefaultUserAgent, settings.userAgent)

	if assert.NotNil(t, settings.credentials) {
		assert.IsType(t, &NoOpCredentials{}, settings.credentials)
	}
}

func TestNewDefaultSettingsWithOpts(t *testing.T) {
	baseURL := "https://url.com"
	userAgent := "testclient/1.0.0"
	username := "user"
	password := "password"

	settings := NewDefaultSettings(
		WithBaseURL(baseURL),
		WithCredentialsProvider(NewPasswordCredentials(username, password)))

	assert.Equal(t, baseURL, settings.baseURL)
	assert.Equal(t, DefaultUserAgent, settings.userAgent)

	if assert.NotNil(t, settings.credentials) {
		assert.IsType(t, &PasswordCredentials{}, settings.credentials)
	}

	settings = NewDefaultSettings(WithUserAgent(userAgent))

	assert.Equal(t, DefaultBaseURL, settings.baseURL)
	assert.Equal(t, userAgent, settings.userAgent)

	if assert.NotNil(t, settings.credentials) {
		assert.IsType(t, &NoOpCredentials{}, settings.credentials)
	}
}
