package skytap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	v := ""

	assert.Equal(t, v, ptrToStr(&v))
}

func TestStringWithNil(t *testing.T) {
	assert.Equal(t, "", ptrToStr(nil))
}

func TestStringPtr(t *testing.T) {
	v := ""
	assert.Equal(t, v, *strToPtr(v))
}

func TestInt(t *testing.T) {
	v := 1

	assert.Equal(t, v, ptrToInt(&v))
}

func TestIntWithNil(t *testing.T) {
	assert.Equal(t, 0, ptrToInt(nil))
}

func TestIntPtr(t *testing.T) {
	v := 2
	assert.Equal(t, v, *intToPtr(v))
}
