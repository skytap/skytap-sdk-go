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

func TestInt(t *testing.T) {
	v := 1

	assert.Equal(t, v, Int(&v))
}

func TestIntWithNil(t *testing.T) {
	assert.Equal(t, 0, Int(nil))
}

func TestIntPtr(t *testing.T) {
	v := 2
	assert.Equal(t, v, *IntPtr(v))
}
