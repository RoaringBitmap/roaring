package roaring

import "testing"

type bitmapOrBulkMergeFixture struct {
	lefts       [2]*Bitmap
	rights      [2]*Bitmap
	cardinality uint64
}

func newBitmapOrBulkMergeFixture(leftKeys, rightKeys []uint16, copyOnWrite bool) bitmapOrBulkMergeFixture {
	fixture := bitmapOrBulkMergeFixture{}
	for variant := range fixture.lefts {
		left := NewBitmap()
		right := NewBitmap()
		leftLow := uint16(variant * 2)
		rightLow := leftLow + 1
		for _, key := range leftKeys {
			left.Add(uint32(key)<<16 | uint32(leftLow))
		}
		for _, key := range rightKeys {
			right.Add(uint32(key)<<16 | uint32(rightLow))
		}
		if copyOnWrite {
			left.SetCopyOnWrite(true)
			right.SetCopyOnWrite(true)
		}
		fixture.lefts[variant] = left
		fixture.rights[variant] = right
	}
	fixture.cardinality = bitmapOrBulkMergeExpected(fixture.lefts[0], fixture.rights[0]).GetCardinality()
	return fixture
}

func bitmapOrBulkMergeExpected(left, right *Bitmap) *Bitmap {
	values := make([]uint32, 0, left.GetCardinality()+right.GetCardinality())
	values = append(values, left.ToArray()...)
	values = append(values, right.ToArray()...)
	return BitmapOf(values...)
}

func bitmapOrBulkMergeKeys(start, count, step int) []uint16 {
	keys := make([]uint16, count)
	for i := range keys {
		keys[i] = uint16(start + i*step)
	}
	return keys
}

func bitmapOrBulkMergeInterleavedFixture(containers int, copyOnWrite bool) bitmapOrBulkMergeFixture {
	return newBitmapOrBulkMergeFixture(
		bitmapOrBulkMergeKeys(0, containers, 2),
		bitmapOrBulkMergeKeys(1, containers, 2),
		copyOnWrite,
	)
}

func bitmapOrBulkMergeAppendFixture(containers int) bitmapOrBulkMergeFixture {
	return newBitmapOrBulkMergeFixture(
		bitmapOrBulkMergeKeys(0, containers, 1),
		bitmapOrBulkMergeKeys(containers, containers, 1),
		false,
	)
}

func bitmapOrBulkMergeOverlapFixture(containers int) bitmapOrBulkMergeFixture {
	keys := bitmapOrBulkMergeKeys(0, containers, 1)
	return newBitmapOrBulkMergeFixture(keys, keys, false)
}

func bitmapOrBulkMergeSingleInteriorFixture(containers int) bitmapOrBulkMergeFixture {
	leftKeys := make([]uint16, 0, containers-1)
	middle := containers / 2
	for key := 0; key < containers; key++ {
		if key != middle {
			leftKeys = append(leftKeys, uint16(key))
		}
	}
	return newBitmapOrBulkMergeFixture(leftKeys, []uint16{uint16(middle)}, false)
}

func bitmapOrBulkMergeFixtureCases() map[string]bitmapOrBulkMergeFixture {
	return map[string]bitmapOrBulkMergeFixture{
		"interleaved-64":                 bitmapOrBulkMergeInterleavedFixture(64, false),
		"interleaved-65":                 bitmapOrBulkMergeInterleavedFixture(65, false),
		"interleaved-1024":               bitmapOrBulkMergeInterleavedFixture(1024, false),
		"interleaved-4096":               bitmapOrBulkMergeInterleavedFixture(4096, false),
		"append-only-4096":               bitmapOrBulkMergeAppendFixture(4096),
		"overlap-4096":                   bitmapOrBulkMergeOverlapFixture(4096),
		"copy-on-write-interleaved-4096": bitmapOrBulkMergeInterleavedFixture(4096, true),
		"single-interior-4096":           bitmapOrBulkMergeSingleInteriorFixture(4096),
		"mixed-key-order": newBitmapOrBulkMergeFixture(
			[]uint16{2, 4, 6, 8},
			[]uint16{1, 4, 5, 8, 9},
			false,
		),
	}
}

func TestBitmapOrBulkMergeFixtures(t *testing.T) {
	for name, fixture := range bitmapOrBulkMergeFixtureCases() {
		t.Run(name, func(t *testing.T) {
			left := fixture.lefts[0]
			right := fixture.rights[0]
			want := bitmapOrBulkMergeExpected(left, right)
			receiver := left.Clone()
			receiver.Or(right)

			if !receiver.Equals(want) {
				t.Fatalf("unexpected union: got %v, want %v", receiver, want)
			}
			if receiver.GetCardinality() != fixture.cardinality {
				t.Fatalf("unexpected cardinality: got %d, want %d", receiver.GetCardinality(), fixture.cardinality)
			}
			if err := receiver.Validate(); err != nil {
				t.Fatalf("union produced an invalid bitmap: %v", err)
			}
		})
	}
}

func TestBitmapOrBulkMergeCopyOnWriteOwnership(t *testing.T) {
	fixture := bitmapOrBulkMergeInterleavedFixture(64, true)
	left := fixture.lefts[0]
	right := fixture.rights[0]
	receiver := left.Clone()
	receiver.Or(right)

	const (
		leftKey   = uint32(0) << 16
		rightKey  = uint32(1) << 16
		receiver1 = uint32(10)
		receiver2 = uint32(11)
	)

	receiver.Add(rightKey | receiver1)
	if right.Contains(rightKey | receiver1) {
		t.Fatal("receiver mutation changed a source-only container")
	}
	right.Add(rightKey | receiver2)
	if receiver.Contains(rightKey | receiver2) {
		t.Fatal("source mutation changed a receiver source-only container")
	}

	receiver.Add(leftKey | receiver1)
	if left.Contains(leftKey | receiver1) {
		t.Fatal("receiver mutation changed a receiver-only container")
	}
	left.Add(leftKey | receiver2)
	if receiver.Contains(leftKey | receiver2) {
		t.Fatal("left mutation changed a receiver container")
	}

	if err := receiver.Validate(); err != nil {
		t.Fatalf("receiver became invalid after copy-on-write mutations: %v", err)
	}
	if err := left.Validate(); err != nil {
		t.Fatalf("left became invalid after copy-on-write mutations: %v", err)
	}
	if err := right.Validate(); err != nil {
		t.Fatalf("right became invalid after copy-on-write mutations: %v", err)
	}
}

func BenchmarkBitmapOrBulkMerge(b *testing.B) {
	fixtures := []struct {
		name    string
		fixture bitmapOrBulkMergeFixture
	}{
		{"fresh-interleaved-64", bitmapOrBulkMergeInterleavedFixture(64, false)},
		{"fresh-interleaved-65", bitmapOrBulkMergeInterleavedFixture(65, false)},
		{"fresh-interleaved-1024", bitmapOrBulkMergeInterleavedFixture(1024, false)},
		{"fresh-interleaved-4096", bitmapOrBulkMergeInterleavedFixture(4096, false)},
		{"fresh-append-only-4096", bitmapOrBulkMergeAppendFixture(4096)},
		{"fresh-overlap-4096", bitmapOrBulkMergeOverlapFixture(4096)},
		{"fresh-copy-on-write-interleaved-4096", bitmapOrBulkMergeInterleavedFixture(4096, true)},
		{"fresh-single-interior-4096", bitmapOrBulkMergeSingleInteriorFixture(4096)},
	}
	for _, benchmark := range fixtures {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			var cardinality uint64
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fixtureIndex := i & 1
				receiver := benchmark.fixture.lefts[fixtureIndex].Clone()
				receiver.Or(benchmark.fixture.rights[fixtureIndex])
				cardinality += receiver.GetCardinality()
			}
			b.StopTimer()
			if cardinality != benchmark.fixture.cardinality*uint64(b.N) {
				b.Fatalf("unexpected total cardinality: got %d, want %d", cardinality, benchmark.fixture.cardinality*uint64(b.N))
			}
		})
	}
}
