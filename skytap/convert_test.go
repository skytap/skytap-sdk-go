package skytap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	v := ""

	assert.Equal(t, v, String(&v))
}

func TestStringWithNil(t *testing.T) {
	assert.Equal(t, "", String(nil))
}

func TestStringPtr(t *testing.T) {
	v := ""
	assert.Equal(t, v, *StringPtr(v))
}
