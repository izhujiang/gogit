package common

import (
	"testing"
)

func TestFunction(t *testing.T) {
	pairA := []*NameHashPair{
		{Oid: hash("2ac3f27df406e454dd936cf4b8fa1645c080e25f"), Name: "file/README.md"},
		{Oid: hash("43758f2e04f8c82080f3cc23a2570de40e57c1bd"), Name: "file/common/hash.go"},
		{Oid: hash("571a0801da171bb450ab69fef9f74fe5b82e4a98"), Name: "file/core/index/index.go"},
		{Oid: hash("69e077dcee8f9a6e7abdb416677016ef9f17e0c1"), Name: "file/core/object.go"},
		{Oid: hash("7d69ffa6a5bea7295324dc8ece14ce04676f259d"), Name: "file/core/stagingarea.go"},
		{Oid: hash("460e47dc5574eab0e6b52bb912b4f4f5f8a9c15e"), Name: "file/core/tree.go"},
		{Oid: hash("b848a979616afb05bf412a5b83201406a3409485"), Name: "file/filemode/filemode.go"},
		{Oid: hash("8d713f6f077191db43d8dfb18d72e406a941b3ae"), Name: "file/filemode/filemode_test.go"},
		{Oid: hash("3103d10731945aa5de9233042a1af1b87f0fcd7e"), Name: "file/go.mod"},
		{Oid: hash("64e4b5caefab89028792f5d0e7b08f9f68ec3169"), Name: "file/main.go"},
		{Oid: hash("906616d09fb892918a9ac3a9dcf84b5102dafb3d"), Name: "hello/go.mod"},
		{Oid: hash("21b72b84646f9141928072bbc2daf3859fe14fa1"), Name: "hello/main.go"},
	}
	pairB := []*NameHashPair{
		{Oid: hash("43758f2e04f8c82080f3cc23a2570de40e57c1bd"), Name: "file/common/hash.go"},
		{Oid: hash("571a0801da171bb450ab69fef9f74fe5b82e4a98"), Name: "file/core/index/index.go"},
		{Oid: hash("69e077dcee8f9a6e7abdb416677016ef9f17e0c1"), Name: "file/core/object.go"},
		{Oid: hash("7d6994a6a5bea7295324dc8ece14ce04676f259d"), Name: "file/core/stagingarea.go"},
		{Oid: hash("ff0e47dc5574eab0e6b52bb912b4f4f5f8a9c15e"), Name: "file/core/tree.go"},
		{Oid: hash("b848a979616afb05bf412a5b83201406a3409485"), Name: "file/filemode/filemode.go"},
		{Oid: hash("8d713f6f077191db43d8dfb18d72e406a941b3ae"), Name: "file/filemode/filemode_test.go"},
		{Oid: hash("64e4b5caefab89028792f5d0e7b08f9f68ec3169"), Name: "file/main.go"},
		{Oid: hash("64b5caefffab89028792f5d0e7b08f9f68ec3169"), Name: "file/main2.go"},
		{Oid: hash("906616d09fb892918a9ac3a9dcf84b5102dafb3d"), Name: "hello/go.mod"},
		{Oid: hash("aff8ba2a067eb5ac9729a5a922b8bb6fe64686c6"), Name: "hello/go.sum"},
	}
	changes := CompareOrderedNameHashPairs(pairA, pairB)

	// assert(t, []string{"file/core/stagingarea.go", "file/core/tree.go"}, modified)
	// assert(t, []string{"file/README.md", "file/go.mod", "hello/main.go"}, removed)
	// assert(t, []string{"file/main2.go", "hello/go.sum"}, added)
	want := make([]*Change, 2)
	want[0] = &Change{nil, pairB[8]}
	want[1] = &Change{nil, pairB[10]}
	assertChanges(t, want, changes.Create)

	want = make([]*Change, 3)
	want[0] = &Change{pairA[0], nil}
	want[1] = &Change{pairA[8], nil}
	want[2] = &Change{pairA[11], nil}
	assertChanges(t, want, changes.Remove)

	want = make([]*Change, 2)
	want[0] = &Change{pairA[4], pairB[3]}
	want[1] = &Change{pairA[5], pairB[4]}
	assertChanges(t, want, changes.Modify)

}

func hash(oid string) Hash {
	h, _ := NewHash(oid)
	return h
}

func assertChanges(t *testing.T, expects []*Change, gots []*Change) {
	if len(expects) != len(gots) {
		t.Error("expect's length != got's length ")

	}
	for i, e := range expects {
		g := gots[i]
		if e.From != g.From || e.To != g.To {
			t.Error("expect != got ", e, g)
		}

	}

}
