package roaring

import "testing"

func TestBitmapOrBulkMergeCopyOnWriteTailOwnership(t *testing.T) {
	fixture := bitmapOrBulkMergeInterleavedFixture(64, true)
	left := fixture.lefts[0]
	right := fixture.rights[0]
	receiver := left.Clone()
	receiver.Or(right)

	tailIndex := right.highlowcontainer.size() - 1
	tailKey := right.highlowcontainer.getKeyAtIndex(tailIndex)
	receiverTailIndex := receiver.highlowcontainer.getIndex(tailKey)
	if receiverTailIndex < 0 {
		t.Fatal("receiver is missing the source-only tail container")
	}
	if !right.highlowcontainer.needsCopyOnWrite(tailIndex) {
		t.Fatal("source-only tail container was not marked copy-on-write")
	}
	if !receiver.highlowcontainer.needsCopyOnWrite(receiverTailIndex) {
		t.Fatal("receiver tail container was not marked copy-on-write")
	}
	if receiver.highlowcontainer.getContainerAtIndex(receiverTailIndex) != right.highlowcontainer.getContainerAtIndex(tailIndex) {
		t.Fatal("source-only tail container was not shared")
	}

	receiverValue := uint32(tailKey)<<16 | 10
	sourceValue := uint32(tailKey)<<16 | 11
	receiver.Add(receiverValue)
	if right.Contains(receiverValue) {
		t.Fatal("receiver tail mutation changed the source")
	}
	right.Add(sourceValue)
	if receiver.Contains(sourceValue) {
		t.Fatal("source tail mutation changed the receiver")
	}

	if err := receiver.Validate(); err != nil {
		t.Fatalf("receiver became invalid after tail mutations: %v", err)
	}
	if err := right.Validate(); err != nil {
		t.Fatalf("source became invalid after tail mutations: %v", err)
	}
}
