package skytap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpCredentials(t *testing.T) {
	cred := NewNoOpCredentials()

	result, err := cred.Retrieve(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestApiTokenCredentials(t *testing.T) {
	username := "user"
	token := "token"
	header := "Basic dXNlcjp0b2tlbg=="

	cred := NewAPITokenCredentials(username, token)

	assert.Equal(t, username, cred.Username)
	assert.Equal(t, token, cred.APIToken)

	result, err := cred.Retrieve(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, header, result)
}
