package skytap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	v := ""

	assert.Equal(t, v, stString(&v))
}

func TestStringWithNil(t *testing.T) {
	assert.Equal(t, "", stString(nil))
}

func TestStringPtr(t *testing.T) {
	v := ""
	assert.Equal(t, v, *stStringPtr(v))
}

func TestInt(t *testing.T) {
	v := 1

	assert.Equal(t, v, stInt(&v))
}

func TestIntWithNil(t *testing.T) {
	assert.Equal(t, 0, stInt(nil))
}

func TestIntPtr(t *testing.T) {
	v := 2
	assert.Equal(t, v, *stIntPtr(v))
}
