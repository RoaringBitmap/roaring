package roaring

import "testing"

// Every in-place set operation must leave its argument untouched, and the
// two bitmaps must not share storage afterwards.
func TestInPlaceOpsPreserveArgument(t *testing.T) {
	const hi = 1 << 16
	dense := func(base uint32) *Bitmap {
		b := New()
		for i := uint32(0); i < 65536; i += 2 {
			b.Add(base | i)
		}
		return b
	}
	run := func(base uint32) *Bitmap {
		b := New()
		b.AddRange(uint64(base)+100, uint64(base)+60000)
		return b
	}
	twoKeys := func(b *Bitmap) *Bitmap {
		b = b.Clone()
		b.AddRange(hi, hi+5)
		return b
	}
	shapes := []struct {
		name string
		bm   *Bitmap
	}{
		{"empty", New()}, {"array", BitmapOf(1, 3, 65535)}, {"bitmap", dense(0)}, {"run", run(0)},
		{"array+key", twoKeys(BitmapOf(2, 4))}, {"run+key", twoKeys(run(0))},
		{"hi-array", BitmapOf(hi|1, hi|3)}, {"hi-bitmap", dense(hi)}, {"hi-run", run(hi)},
	}
	ops := []struct {
		name    string
		inPlace func(a, b *Bitmap)
		pure    func(a, b *Bitmap) *Bitmap
	}{
		{"Or", func(a, b *Bitmap) { a.Or(b) }, Or},
		{"And", func(a, b *Bitmap) { a.And(b) }, And},
		{"AndNot", func(a, b *Bitmap) { a.AndNot(b) }, AndNot},
		{"Xor", func(a, b *Bitmap) { a.Xor(b) }, Xor},
	}
	for _, op := range ops {
		for _, left := range shapes {
			for _, right := range shapes {
				l, r := left.bm, right.bm
				for _, cow := range []bool{false, true} {
					name := op.name + "/" + left.name + "/" + right.name
					a, b := l.Clone(), r.Clone()
					a.SetCopyOnWrite(cow)
					b.SetCopyOnWrite(cow)
					snapshot := a.Clone()
					want := op.pure(l, r)
					op.inPlace(a, b)
					if !snapshot.Equals(l) {
						t.Fatalf("%s: a clone of the receiver changed (copy-on-write %v)", name, cow)
					}
					if !a.Equals(want) {
						t.Fatalf("%s: wrong result", name)
					}
					if err := a.Validate(); err != nil {
						t.Fatalf("%s: %v", name, err)
					}
					if !b.Equals(r) {
						t.Fatalf("%s: argument modified (copy-on-write %v)", name, cow)
					}
					it := r.Iterator()
					for n := 0; n < 50 && it.HasNext(); n++ {
						a.Remove(it.Next())
					}
					a.Add(65533)
					a.AddRange(hi+70000, hi+70050)
					if !b.Equals(r) {
						t.Fatalf("%s: editing the result changed the argument (copy-on-write %v)", name, cow)
					}
				}
			}
			a := left.bm.Clone()
			op.inPlace(a, a)
			if !a.Equals(op.pure(left.bm, left.bm)) {
				t.Fatalf("%s/%s: wrong result for a bitmap applied to itself", op.name, left.name)
			}
		}
	}
}
