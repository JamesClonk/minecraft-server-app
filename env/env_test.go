package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Env_Get(t *testing.T) {
	assert.NotEqual(t, "blubb", Get("PATH", "blubb"))
	assert.Equal(t, "blabb", Get("blibb", "blabb"))
}

func Test_Env_MustGet(t *testing.T) {
	path := MustGet("PATH")
	assert.NotNil(t, path)

	called := false
	defer func() {
		err := recover()
		if err != nil {
			assert.Equal(t, "Required ENV variable [blebb] is missing!", err)
			called = true
		}
	}()
	MustGet("blebb")
	assert.True(t, called)
}
