package domain

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLoadStops(t *testing.T) {
	t.Run("invalid format", func(t *testing.T) {
		stops, err := LoadStops(strings.NewReader("not a valid json"))
		assert.Nil(t, stops, "stops should be nil in case of an err")
		assert.EqualError(t, err, "could not parse stop definition: invalid character 'o' in literal null (expecting 'u')", "err message not correct")
	})
}
