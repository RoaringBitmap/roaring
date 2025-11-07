package roaring

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type bmRec struct {
	Shard       int    `json:"shard"`
	Bucket      int    `json:"bucket"`
	Sid         string `json:"sid"`
	D           string `json:"d"`
	ChunkSum    string `json:"chunk_sum"`
	OriginalSum string `json:"original_sum"`
}

func readAndJoinBm(t *testing.T, filename string, funcOR func(bms ...*Bitmap) *Bitmap) *Bitmap {
	file, err := os.ReadFile(filename)
	require.NoError(t, err)

	var recs []bmRec
	err = json.Unmarshal(file, &recs)
	require.NoError(t, err)

	chunks := make([]*Bitmap, 0)

	for _, rec := range recs {
		rb := NewBitmap()
		_, err = rb.FromBase64(rec.D)
		require.NoErrorf(t, err, "rb.FromBase64, sid: %v", rec.Sid)

		err = rb.Validate()
		if err != nil {
			t.Logf(
				"readAndJoinBm(%v): rb.Validate(%v); card: %v; %v\n",
				filename, rec.Sid, rb.GetCardinality(), err,
			)
		} else {
			t.Logf(
				"readAndJoinBm(%v): rb.Validate(%v); card: %v; %v\n",
				filename, rec.Sid, rb.GetCardinality(), "valid",
			)
		}

		chunks = append(chunks, rb)
	}

	total := funcOR(chunks...)
	err = total.Validate()
	if err != nil {
		t.Logf("readAndJoinBm(%v): total.Validate(): %v\n", filename, err)
	}

	t.Logf("card(%v): %v\n", filename, total.GetCardinality())
	b, err := total.ToBytes()
	if err != nil {
		t.Error("readAndJoinBm : ToBytes error", err.Error())
	} else {
		testrb := NewBitmap()
		err = testrb.UnmarshalBinary(b)
		require.NoErrorf(t, err, "readAndJoinBm: UnmarshalBinary")
	}

	return total
}

const (
	bmVersionFrom213 = "mcp-ss-bitmaps_31_10_25_16_59.json"
)

func TestReadBitmapFromRoaring_v2_13_0_And_MakeOperationBy_v2_6_0(t *testing.T) {
	// read chunks for join to one bitmap via FastOr
	rb := readAndJoinBm(t, bmVersionFrom213, FastOr) // use HeapOr for fixed test. What kind problem with FastOr?

	// init masks
	masks := make([]*Bitmap, 0, 30)
	for i := 0; i < 30; i++ {
		bm := NewBitmap()
		bm.AddRange(uint64((10_000_000*i)+1), uint64(10_000_000*(i+1)))
		masks = append(masks, bm)
	}

	chunks := make([]*Bitmap, 0)
	// split by chunk
	for _, mask := range masks {
		chunkBm := mask.Clone()
		chunkBm.And(rb)

		if !chunkBm.IsEmpty() {
			chunks = append(chunks, chunkBm)
		}
	}

	// Check validate and UnmarshalBinary
	for i, chunk := range chunks {
		// Validate ?
		err := chunk.Validate()
		if err != nil {
			fmt.Printf(
				"readAndJoinBm(%v): rb.Validate(%v); card: %v; %v\n",
				bmVersionFrom213, i, chunk.GetCardinality(), err,
			)
		} else {
			fmt.Printf(
				"readAndJoinBm(%v): rb.Validate(%v); card: %v; %v\n",
				bmVersionFrom213, i, rb.GetCardinality(), "valid",
			)
		}

		// Binary OK ?
		b, err := chunk.ToBytes()
		require.NoError(t, err, "chunksToBytes: conver", i)
		testrb := NewBitmap()
		err = testrb.UnmarshalBinary(b)
		// !!! TEST FAIL HERE !!!
		require.NoErrorf(t, err, "chunksToBytes: UnmarshalBinary, %s: %v", bmVersionFrom213, i)
	}

}
