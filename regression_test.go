package roaring

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOrRegression is a regression test as described in this issue:
// https://github.com/RoaringBitmap/roaring/issues/358
func TestCorruption(t *testing.T) {
	bm1Raw := "OjAAAAEAAAAAAAoAEAAAANeU2JTZlNqU25TclN2U3pTflOCU4ZQ="
	bm2Raw := "OzADAA8AAPT/AQD//wIA//8DAD8NJQAAAC8AAAA1AAAAOwAAAAIAAADWlOKUHWsBAAAA//8BAAAA//8BAAAAPw0="
	bm1Decoded, err := base64.StdEncoding.DecodeString(bm1Raw)
	require.NoError(t, err)
	bm2Decoded, err := base64.StdEncoding.DecodeString(bm2Raw)
	require.NoError(t, err)

	bm1 := New()
	_, err = bm1.FromBuffer(bm1Decoded)
	require.NoError(t, err)

	roundTripRoaring(t, bm1)

	bm2 := New()
	_, err = bm2.FromBuffer(bm2Decoded)
	require.NoError(t, err)

	roundTripRoaring(t, bm2)

	fmt.Println(bm1.String())
	fmt.Println(bm2.String())

	bm1.Or(bm2)

	roundTripRoaring(t, bm1)
}

func roundTripRoaring(t *testing.T, b *Bitmap) {
	marshaled, err := b.ToBytes()
	require.NoError(t, err)
	p, err := New().FromBuffer(marshaled)
	require.NoError(t, err)
	require.Equal(t, int64(len(marshaled)), p)
}
