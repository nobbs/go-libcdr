package libcdr_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/nobbs/go-libcdr"
)

// newConverter builds a Converter bound to the test's lifetime.
func newConverter(t *testing.T) *libcdr.Converter {
	t.Helper()

	converter, err := libcdr.New(t.Context())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	t.Cleanup(func() { _ = converter.Close(context.WithoutCancel(t.Context())) })

	return converter
}

// fixtures returns the .cdr files in testdata. They are not committed: see
// testdata/README.md.
func fixtures(t *testing.T) []string {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join("testdata", "*.cdr"))
	if err != nil {
		t.Fatalf("globbing fixtures: %v", err)
	}

	return matches
}

func TestConvertRejectsBadInput(t *testing.T) {
	t.Parallel()
	converter := newConverter(t)

	tests := []struct {
		name  string
		input []byte
		want  error
	}{
		{"empty", nil, libcdr.ErrUnsupportedDocument},
		{"plain text", []byte("this is not a CorelDRAW file"), libcdr.ErrUnsupportedDocument},
		{"riff header only", []byte("RIFF\x00\x00\x00\x00CDR"), libcdr.ErrUnsupportedDocument},
		{"truncated riff", append([]byte("RIFF"), bytes.Repeat([]byte{0xff}, 64)...), libcdr.ErrUnsupportedDocument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := converter.Convert(t.Context(), tt.input)
			if err == nil {
				t.Fatal("Convert() succeeded, want an error")
			}

			if !errors.Is(err, tt.want) {
				t.Errorf("Convert() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestConvertFixtures(t *testing.T) {
	t.Parallel()

	paths := fixtures(t)
	if len(paths) == 0 {
		t.Skip("no .cdr fixtures in testdata; see testdata/README.md")
	}

	converter := newConverter(t)

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			document, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading fixture: %v", err)
			}

			svg, err := converter.Convert(t.Context(), document)
			if err != nil {
				t.Fatalf("Convert: %v", err)
			}

			if !bytes.Contains(svg, []byte("<svg")) {
				t.Errorf("output is not SVG (%d bytes)", len(svg))
			}

			// Conversion must be deterministic: the same input always yields
			// byte-identical output.
			again, err := converter.Convert(t.Context(), document)
			if err != nil {
				t.Fatalf("second Convert: %v", err)
			}

			if !bytes.Equal(svg, again) {
				t.Error("Convert() is not deterministic across calls")
			}
		})
	}
}

func TestConvertIsConcurrencySafe(t *testing.T) {
	t.Parallel()

	paths := fixtures(t)
	if len(paths) == 0 {
		t.Skip("no .cdr fixtures in testdata; see testdata/README.md")
	}

	document, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	converter := newConverter(t)

	want, err := converter.Convert(t.Context(), document)
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}

	var wg sync.WaitGroup

	errs := make([]error, 4)

	results := make([][]byte, 4)
	for i := range results {
		wg.Go(func() {
			results[i], errs[i] = converter.Convert(t.Context(), document)
		})
	}

	wg.Wait()

	for i := range results {
		if errs[i] != nil {
			t.Errorf("concurrent Convert %d: %v", i, errs[i])
			continue
		}

		if !bytes.Equal(results[i], want) {
			t.Errorf("concurrent Convert %d returned different output", i)
		}
	}
}

func TestConvertHonoursCancellation(t *testing.T) {
	t.Parallel()
	converter := newConverter(t)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := converter.Convert(ctx, []byte("RIFF placeholder document"))
	if err == nil {
		t.Fatal("Convert() succeeded with a cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Convert() error = %v, want context.Canceled", err)
	}
}
