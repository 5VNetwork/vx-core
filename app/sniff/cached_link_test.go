package sniff

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/pipe"

	"github.com/google/go-cmp/cmp"
)

func TestCachedLink(t *testing.T) {
	iLink, oLink := pipe.NewLinks(-1, false)
	r := &CachedReadRw{
		ReaderWriter: oLink,
	}

	b1 := make([]byte, 100)
	rand.Read(b1)

	iLink.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes(b1)})

	b2 := make([]byte, 8192)
	copied, n, err := r.read(b2)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}
	if d := cmp.Diff(b1, b2[:n]); d != "" {
		t.Error(d)
	}

	mb, err := r.ReadMultiBuffer()
	if err != nil {
		t.Error(err)
	}
	if d := cmp.Diff(b1, mb[0].Bytes()); d != "" {
		t.Error(d)
	}

	for i := 0; i < 5; i++ {
		rand.Read(b1)
		iLink.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes(b1)})
		mb, err := r.ReadMultiBuffer()
		if err != nil {
			t.Error(err)
		}
		if d := cmp.Diff(b1, mb[0].Bytes()); d != "" {
			t.Error(d)
		}
	}
}

func TestCachedReadRwBasic(t *testing.T) {
	iLink, oLink := pipe.NewLinks(-1, false)
	r := &CachedReadRw{
		ReaderWriter: oLink,
	}

	b1 := make([]byte, 100)
	rand.Read(b1)

	iLink.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes(b1)})

	b2 := make([]byte, 8192)
	copied, n, err := r.read(b2)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}
	if d := cmp.Diff(b1, b2[:n]); d != "" {
		t.Error(d)
	}

	mb, err := r.ReadMultiBuffer()
	if err != nil {
		t.Error(err)
	}
	if d := cmp.Diff(b1, mb[0].Bytes()); d != "" {
		t.Error(d)
	}
}

func TestCachedReadRwMultipleReads(t *testing.T) {
	iLink, oLink := pipe.NewLinks(-1, false)
	r := &CachedReadRw{
		ReaderWriter: oLink,
	}

	for i := 0; i < 5; i++ {
		testData := make([]byte, 100)
		rand.Read(testData)
		iLink.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes(testData)})
		mb, err := r.ReadMultiBuffer()
		if err != nil {
			t.Fatal(err)
		}
		if d := cmp.Diff(testData, mb[0].Bytes()); d != "" {
			t.Error(d)
		}
	}
}

func TestCachedReadRwCacheTwice(t *testing.T) {
	iLink, oLink := pipe.NewLinks(-1, false)
	r := &CachedReadRw{
		ReaderWriter: oLink,
	}

	b1 := make([]byte, 100)
	rand.Read(b1)

	iLink.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes(b1)})
	iLink.CloseWrite()

	b2 := make([]byte, 8192)
	copied, n1, err := r.read(b2)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}
	if d := cmp.Diff(b1, b2[:n1]); d != "" {
		t.Error(d)
	}

	_, n2, err := r.read(b2)
	if err != io.EOF {
		t.Error(err)
	}
	if n2 != 0 {
		t.Errorf("Expected second cache call to return 0, got %d", n2)
	}

	mb, err := r.ReadMultiBuffer()
	if err != nil {
		t.Error(err)
	}
	if d := cmp.Diff(b1, mb[0].Bytes()); d != "" {
		t.Error(d)
	}
}
