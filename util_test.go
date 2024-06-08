package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLobHob(t *testing.T) {
	for i := 0; i < 2049; i++ {
		val := uint32(i)
		lob := lowbits(uint32(val))
		hob := highbits(uint32(val))
		reconstructed := combineLoHi16(lob, hob)
		assert.Equal(t, reconstructed, val)
	}

	for i := 0; i < 2049; i++ {
		val := uint32(i)
		lob := lowbits(uint32(val))
		hob := highbits(uint32(val))
		reconstructed := combineLoHi32(uint32(lob), uint32(hob))
		assert.Equal(t, reconstructed, val)
	}
}
