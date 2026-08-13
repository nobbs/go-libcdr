package libcdr

import (
	"context"
	"errors"
	"testing"

	"github.com/tetratelabs/wazero/sys"
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

func TestConvertErrorExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		exitCode uint32
		detail   string
		want     error
		message  string
	}{
		{
			name:     "unsupported without detail",
			exitCode: exitUnsupported,
			want:     ErrUnsupportedDocument,
			message:  ErrUnsupportedDocument.Error(),
		},
		{
			name:     "parse failure without detail",
			exitCode: exitParseFailed,
			want:     ErrParseFailed,
			message:  ErrParseFailed.Error(),
		},
		{
			name:     "unsupported with detail",
			exitCode: exitUnsupported,
			detail:   "unsupported format",
			want:     ErrUnsupportedDocument,
			message:  "libcdr: unsupported document: unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := convertError(
				context.Background(),
				sys.NewExitError(tt.exitCode),
				[]byte(tt.detail),
			)

			if !errors.Is(err, tt.want) {
				t.Errorf("convertError() error = %v, want %v", err, tt.want)
			}

			if err.Error() != tt.message {
				t.Errorf("convertError() error = %q, want %q", err, tt.message)
			}
		})
	}
}
