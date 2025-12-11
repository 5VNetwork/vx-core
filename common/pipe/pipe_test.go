package pipe

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
)

func TestPipeReadWrite(t *testing.T) {
	pipe := NewPipe(1024, true)

	b := buf.New()
	b.WriteString("abcd")
	common.Must(pipe.WriteMultiBuffer(buf.MultiBuffer{b}))

	b2 := buf.New()
	b2.WriteString("efg")
	common.Must(pipe.WriteMultiBuffer(buf.MultiBuffer{b2}))

	rb, err := pipe.ReadMultiBuffer()
	common.Must(err)
	if r := cmp.Diff(rb.String(), "abcdefg"); r != "" {
		t.Error(r)
	}
}

func TestPipeInterrupt(t *testing.T) {
	pipe := NewPipe(1024, true)
	payload := []byte{'a', 'b', 'c', 'd'}
	b := buf.New()
	b.Write(payload)
	pipe.WriteMultiBuffer(buf.MultiBuffer{b})
	pipe.Interrupt(nil)

	rb, err := pipe.ReadMultiBuffer()
	if !errors.Is(err, errPipeInterrupted) {
		t.Fatal("expect io.ErrClosePipe, but got ", err)
	}
	if !rb.IsEmpty() {
		t.Fatal("expect empty buffer, but got ", rb.Len())
	}
}

func TestPipeClose(t *testing.T) {
	pipe := NewPipe(1024, true)

	payload := []byte{'a', 'b', 'c', 'd'}
	b := buf.New()
	b.Write(payload)
	pipe.WriteMultiBuffer(buf.MultiBuffer{b})
	pipe.Close()

	rb, err := pipe.ReadMultiBuffer()
	common.Must(err)
	if rb.String() != string(payload) {
		t.Fatal("expect content ", string(payload), " but actually ", rb.String())
	}

	rb, err = pipe.ReadMultiBuffer()
	if err != io.EOF {
		t.Fatal("expected EOF, but got ", err)
	}
	if !rb.IsEmpty() {
		t.Fatal("expect empty buffer, but got ", rb.String())
	}
}

func TestPipeLimitZero(t *testing.T) {
	pipe := NewPipe(0, false)
	bb := buf.New()
	bb.Write([]byte{'a', 'b'})
	pipe.WriteMultiBuffer(buf.MultiBuffer{bb})

	var errg errgroup.Group
	errg.Go(func() error {
		b := buf.New()
		b.Write([]byte{'c', 'd'})
		return pipe.WriteMultiBuffer(buf.MultiBuffer{b})
	})
	errg.Go(func() error {
		time.Sleep(time.Second)

		var container buf.MultiBufferContainer
		if err := buf.Copy(pipe, &container); err != nil {
			return err
		}

		if r := cmp.Diff(container.String(), "abcd"); r != "" {
			return errors.New(r)
		}
		return nil
	})
	errg.Go(func() error {
		time.Sleep(time.Second * 2)
		pipe.Close()
		return nil
	})
	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestPipeWriteMultiThread(t *testing.T) {
	p := NewPipe(0, false)

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(func() error {
			b := buf.New()
			b.WriteString("abcd")
			return p.WriteMultiBuffer(buf.MultiBuffer{b})
		})
	}
	time.Sleep(time.Millisecond * 100)
	p.Close()
	errg.Wait()

	b, err := p.ReadMultiBuffer()
	common.Must(err)
	if r := cmp.Diff(b[0].Bytes(), []byte{'a', 'b', 'c', 'd'}); r != "" {
		t.Error(r)
	}
}

func BenchmarkPipeReadWrite(b *testing.B) {
	p := NewPipe(-1, false)
	a := buf.New()
	a.Extend(buf.Size)
	c := buf.MultiBuffer{a}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		common.Must(p.WriteMultiBuffer(c))
		d, err := p.ReadMultiBuffer()
		common.Must(err)
		c = d
	}
}

func TestPipeReadDeadline(t *testing.T) {
	pipe := NewPipe(1024, false)

	// Set read deadline in the near future
	pipe.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	// Try to read from empty pipe - should timeout
	start := time.Now()
	_, err := pipe.ReadMultiBuffer()
	elapsed := time.Since(start)

	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrReadTimeout, got %v", err)
	}

	// Should timeout around 100ms
	if elapsed < 50*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("timeout took %v, expected around 100ms", elapsed)
	}
}

func TestPipeReadDeadlineAlreadyExpired(t *testing.T) {
	pipe := NewPipe(1024, false)

	// Set read deadline in the past
	pipe.SetReadDeadline(time.Now().Add(-1 * time.Second))

	// Try to read - should timeout immediately
	start := time.Now()
	_, err := pipe.ReadMultiBuffer()
	elapsed := time.Since(start)

	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrReadTimeout, got %v", err)
	}

	// Should timeout immediately
	if elapsed > 50*time.Millisecond {
		t.Errorf("timeout took %v, expected immediate timeout", elapsed)
	}
}

func TestPipeReadDeadlineReset(t *testing.T) {
	pipe := NewPipe(1024, false)

	// Set a short read deadline
	pipe.SetReadDeadline(time.Now().Add(50 * time.Millisecond))

	// Reset deadline to zero (no deadline)
	time.Sleep(20 * time.Millisecond)
	pipe.SetReadDeadline(time.Time{})

	// Write data in background after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		b := buf.New()
		b.WriteString("test")
		pipe.WriteMultiBuffer(buf.MultiBuffer{b})
	}()

	// Should not timeout since deadline was reset
	rb, err := pipe.ReadMultiBuffer()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rb.String() != "test" {
		t.Errorf("expected 'test', got %s", rb.String())
	}
}

func TestPipeWriteDeadline(t *testing.T) {
	pipe := NewPipe(0, false) // Zero limit, second write will block

	// First write succeeds
	b1 := buf.New()
	b1.WriteString("first")
	if err := pipe.WriteMultiBuffer(buf.MultiBuffer{b1}); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Set write deadline
	pipe.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))

	// Second write should block and timeout
	b2 := buf.New()
	b2.WriteString("second")
	start := time.Now()
	err := pipe.WriteMultiBuffer(buf.MultiBuffer{b2})
	elapsed := time.Since(start)

	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrWriteTimeout, got %v", err)
	}

	// Should timeout around 100ms
	if elapsed < 50*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("timeout took %v, expected around 100ms", elapsed)
	}
}

func TestPipeWriteDeadlineAlreadyExpired(t *testing.T) {
	pipe := NewPipe(1024, false)

	// Set write deadline in the past
	pipe.SetWriteDeadline(time.Now().Add(-1 * time.Second))

	// Try to write - should timeout immediately
	b := buf.New()
	b.WriteString("test")
	start := time.Now()
	err := pipe.WriteMultiBuffer(buf.MultiBuffer{b})
	elapsed := time.Since(start)

	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrWriteTimeout, got %v", err)
	}

	// Should timeout immediately
	if elapsed > 50*time.Millisecond {
		t.Errorf("timeout took %v, expected immediate timeout", elapsed)
	}
}

func TestPipeWriteDeadlineReset(t *testing.T) {
	pipe := NewPipe(0, false) // Zero limit

	// First write succeeds
	b1 := buf.New()
	b1.WriteString("first")
	if err := pipe.WriteMultiBuffer(buf.MultiBuffer{b1}); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Set a short write deadline
	pipe.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))

	// Reset deadline to zero (no deadline) before it expires
	time.Sleep(20 * time.Millisecond)
	pipe.SetWriteDeadline(time.Time{})

	// Read data in background after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		pipe.ReadMultiBuffer()
	}()

	// Write should succeed since deadline was reset
	b2 := buf.New()
	b2.WriteString("second")
	if err := pipe.WriteMultiBuffer(buf.MultiBuffer{b2}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPipeSetDeadline(t *testing.T) {
	pipe := NewPipe(0, false)

	// First write succeeds
	b1 := buf.New()
	b1.WriteString("data")
	if err := pipe.WriteMultiBuffer(buf.MultiBuffer{b1}); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Set both deadlines using SetDeadline
	deadline := time.Now().Add(100 * time.Millisecond)
	pipe.SetDeadline(deadline)

	// Both read and write should timeout
	var errg errgroup.Group

	errg.Go(func() error {
		b := buf.New()
		b.WriteString("test")
		err := pipe.WriteMultiBuffer(buf.MultiBuffer{b})
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			return errors.New("write: expected ErrWriteTimeout")
		}
		return nil
	})

	// Wait a bit to ensure write blocks first
	time.Sleep(20 * time.Millisecond)

	// Create a new pipe for read test since the first one is blocked
	pipe2 := NewPipe(1024, false)
	pipe2.SetDeadline(time.Now().Add(100 * time.Millisecond))

	errg.Go(func() error {
		_, err := pipe2.ReadMultiBuffer()
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			return errors.New("read: expected os.ErrDeadlineExceeded")
		}
		return nil
	})

	if err := errg.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestPipeReadDeadlineWithData(t *testing.T) {
	pipe := NewPipe(1024, false)

	// Write data first
	b := buf.New()
	b.WriteString("hello")
	if err := pipe.WriteMultiBuffer(buf.MultiBuffer{b}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Set read deadline
	pipe.SetReadDeadline(time.Now().Add(1 * time.Second))

	// Read should succeed immediately since data is available
	rb, err := pipe.ReadMultiBuffer()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rb.String() != "hello" {
		t.Errorf("expected 'hello', got %s", rb.String())
	}
}

func TestPipeWriteDeadlineWithSpace(t *testing.T) {
	pipe := NewPipe(1024, false)

	// Set write deadline
	pipe.SetWriteDeadline(time.Now().Add(1 * time.Second))

	// Write should succeed immediately since there's space
	b := buf.New()
	b.WriteString("hello")
	if err := pipe.WriteMultiBuffer(buf.MultiBuffer{b}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPipeDeadlineConcurrent(t *testing.T) {
	pipe := NewPipe(1024, false)

	var errg errgroup.Group

	// Multiple goroutines setting deadlines concurrently
	for i := 0; i < 10; i++ {
		errg.Go(func() error {
			pipe.SetReadDeadline(time.Now().Add(1 * time.Second))
			pipe.SetWriteDeadline(time.Now().Add(1 * time.Second))
			pipe.SetDeadline(time.Now().Add(1 * time.Second))
			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		t.Fatal(err)
	}

	// Pipe should still be functional
	b := buf.New()
	b.WriteString("test")
	if err := pipe.WriteMultiBuffer(buf.MultiBuffer{b}); err != nil {
		t.Fatalf("write failed after concurrent deadline sets: %v", err)
	}

	rb, err := pipe.ReadMultiBuffer()
	if err != nil {
		t.Fatalf("read failed after concurrent deadline sets: %v", err)
	}
	if rb.String() != "test" {
		t.Errorf("expected 'test', got %s", rb.String())
	}
}
