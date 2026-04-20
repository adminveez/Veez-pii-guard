package pii

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
)

// StreamOptions configures the stream scanner.
type StreamOptions struct {
	// ChunkSize is the target read size in bytes (default 2048).
	ChunkSize int
	// Overlap is the byte length of the trailing window kept between chunks
	// to detect PII straddling chunk boundaries (default 128).
	// MUST be greater than the longest PII pattern you expect to detect.
	Overlap int
	// Detector is the underlying scanner. If nil, a default-config detector is used.
	Detector *Detector
}

// Chunk is one streamed scan result. Offset is the absolute byte offset of
// AnonymizedText within the original stream.
type Chunk struct {
	AnonymizedText string
	Detections     []Detection
	Offset         int64
}

// StreamScanner scans an io.Reader chunk by chunk with a tail-overlap window
// so that PII straddling chunk boundaries is still detected. See ADR-005.
type StreamScanner struct {
	opts     StreamOptions
	detector *Detector
}

// NewStreamScanner constructs a scanner. See StreamOptions for defaults.
func NewStreamScanner(opts StreamOptions) (*StreamScanner, error) {
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = 2048
	}
	if opts.Overlap < 0 {
		return nil, fmt.Errorf("%w: Overlap must be >= 0", ErrInvalidConfig)
	}
	if opts.Overlap == 0 {
		opts.Overlap = 128
	}
	if opts.Overlap >= opts.ChunkSize {
		return nil, fmt.Errorf("%w: Overlap must be < ChunkSize", ErrInvalidConfig)
	}
	if opts.Detector == nil {
		d, err := NewDetector(DefaultConfig())
		if err != nil {
			return nil, err
		}
		opts.Detector = d
	}
	return &StreamScanner{opts: opts, detector: opts.Detector}, nil
}

// Scan reads from r and returns a channel of Chunks. Closes the channel on
// EOF or context cancellation. Errors are reported via the second channel.
func (s *StreamScanner) Scan(ctx context.Context, r io.Reader) (<-chan Chunk, <-chan error) {
	out := make(chan Chunk, 4)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		br := bufio.NewReaderSize(r, s.opts.ChunkSize)
		buf := make([]byte, s.opts.ChunkSize)
		var carry []byte
		var absoluteOffset int64
		seen := map[chunkSeenKey]struct{}{}

		for {
			if err := ctx.Err(); err != nil {
				errCh <- err
				return
			}
			n, err := br.Read(buf)
			if n > 0 {
				combined := append([]byte{}, carry...)
				combined = append(combined, buf[:n]...)
				chunkOffset := absoluteOffset - int64(len(carry))

				res := s.detector.Scan(ctx, string(combined))

				// Filter detections we already emitted in the previous overlap.
				kept := make([]Detection, 0, len(res.Detections))
				for _, det := range res.Detections {
					k := chunkSeenKey{Offset: chunkOffset + int64(det.Start), Type: det.Type, Text: det.Text}
					if _, dup := seen[k]; dup {
						continue
					}
					seen[k] = struct{}{}
					kept = append(kept, det)
				}

				out <- Chunk{
					AnonymizedText: res.AnonymizedText,
					Detections:     kept,
					Offset:         chunkOffset,
				}

				absoluteOffset += int64(n)

				// Carry the tail for the next chunk.
				if len(combined) > s.opts.Overlap {
					carry = append(carry[:0], combined[len(combined)-s.opts.Overlap:]...)
				} else {
					carry = append(carry[:0], combined...)
				}
			}
			if err == io.EOF {
				return
			}
			if err != nil && !errors.Is(err, io.EOF) {
				errCh <- err
				return
			}
		}
	}()

	return out, errCh
}

type chunkSeenKey struct {
	Offset int64
	Type   Type
	Text   string
}
