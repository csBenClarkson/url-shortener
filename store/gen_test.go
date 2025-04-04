package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeToString(t *testing.T) {
	cases := map[uint64]string{
		2240606794:          "2rDmbg",
		224060679442554121:  "gycpUsXcQp",
		1844674407333351615: "2cgCXoRS2Xd",
	}
	for k, v := range cases {
		assert.Equal(t, encodeToString(k), v)
	}
}
