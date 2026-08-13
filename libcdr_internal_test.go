package libcdr

import (
	"errors"
	"testing"
)

func TestLimitedBuffer(t *testing.T) {
	t.Parallel()

	buffer := limitedBuffer{limit: 4}

	if _, err := buffer.Write([]byte("test")); err != nil {
		t.Fatalf("Write within limit: %v", err)
	}

	if _, err := buffer.Write([]byte("!")); !errors.Is(err, ErrOutputTooLarge) {
		t.Fatalf("Write above limit error = %v, want ErrOutputTooLarge", err)
	}

	if got := buffer.buffer.String(); got != "test" {
		t.Errorf("buffer contents = %q, want %q", got, "test")
	}
}

func TestTruncatedBuffer(t *testing.T) {
	t.Parallel()

	buffer := truncatedBuffer{limit: 4}

	n, err := buffer.Write([]byte("testing"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if n != len("testing") {
		t.Errorf("Write count = %d, want %d", n, len("testing"))
	}

	if got := string(buffer.Bytes()); got != "test" {
		t.Errorf("buffer contents = %q, want %q", got, "test")
	}
}
