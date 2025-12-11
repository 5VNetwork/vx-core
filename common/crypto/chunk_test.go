package crypto

import (
	"bytes"
	"io"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
)

func TestChunkStreamIO(t *testing.T) {
	cache := bytes.NewBuffer(make([]byte, 0, 8192))

	writer := NewChunkStreamWriter(PlainChunkSizeParser{}, cache)
	reader := NewChunkStreamReader(PlainChunkSizeParser{}, cache)

	b := buf.New()
	b.WriteString("abcd")
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{b}))

	b = buf.New()
	b.WriteString("efg")
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{b}))

	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{}))

	if cache.Len() != 13 {
		t.Fatalf("Cache length is %d, want 13", cache.Len())
	}

	mb, err := reader.ReadMultiBuffer()
	common.Must(err)

	if s := mb.String(); s != "abcd" {
		t.Error("content: ", s)
	}

	mb, err = reader.ReadMultiBuffer()
	common.Must(err)

	if s := mb.String(); s != "efg" {
		t.Error("content: ", s)
	}

	_, err = reader.ReadMultiBuffer()
	if err != io.EOF {
		t.Error("error: ", err)
	}
}

func TestChunkStreamIO1(t *testing.T) {
	cache := bytes.NewBuffer(make([]byte, 0, 8192))

	writer := NewChunkStreamWriter1(PlainChunkSizeParser{}, cache)
	reader := NewChunkStreamReader1(PlainChunkSizeParser{}, cache, 0)

	b := buf.New()
	b.WriteString("abcd")
	common.Must2(writer.Write(b.Bytes()))

	b = buf.New()
	b.WriteString("efg")
	common.Must2(writer.Write(b.Bytes()))

	common.Must2(writer.Write([]byte{}))

	if cache.Len() != 13 {
		t.Fatalf("Cache length is %d, want 13", cache.Len())
	}

	buffer := make([]byte, 2048)
	mb, err := reader.Read(buffer)
	common.Must(err)

	if s := string(buffer[:mb]); s != "abcd" {
		t.Error("content: ", s)
	}

	mb, err = reader.Read(buffer)
	common.Must(err)

	if s := string(buffer[:mb]); s != "efg" {
		t.Error("content: ", s)
	}

	_, err = reader.Read(buffer)
	if err != io.EOF {
		t.Error("error: ", err)
	}
}
