package table

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RowStream returns rows in deterministic source order until io.EOF.
type RowStream interface {
	Next(context.Context) (Row, error)
	Close() error
}

type SliceRowStream struct {
	rows   []Row
	index  int
	closed bool
}

func NewSliceRowStream(rows []Row) *SliceRowStream {
	return &SliceRowStream{rows: append([]Row(nil), rows...)}
}

func (stream *SliceRowStream) Next(ctx context.Context) (Row, error) {
	if stream.closed {
		return Row{}, errors.New("row stream is closed")
	}
	if err := ctx.Err(); err != nil {
		return Row{}, err
	}
	if stream.index >= len(stream.rows) {
		return Row{}, io.EOF
	}
	row := stream.rows[stream.index]
	stream.index++
	return row, nil
}

func (stream *SliceRowStream) Close() error {
	stream.closed = true
	return nil
}

// BoundedSpool retains encoded rows only up to memoryLimit and then spills the
// complete ordered stream to one temporary JSON sequence.
type BoundedSpool struct {
	directory   string
	memoryLimit int64
	memoryUsed  int64
	records     [][]byte
	file        *os.File
	writer      *bufio.Writer
	path        string
	sealed      bool
	closed      bool
	active      io.Closer
}

func NewBoundedSpool(directory string, memoryLimit int64) (*BoundedSpool, error) {
	if directory == "" {
		return nil, errors.New("spool directory is required")
	}
	if memoryLimit <= 0 {
		return nil, errors.New("spool memory limit must be greater than zero")
	}
	return &BoundedSpool{directory: directory, memoryLimit: memoryLimit}, nil
}

func (spool *BoundedSpool) Append(row Row) error {
	if spool.closed {
		return errors.New("spool is closed")
	}
	if spool.sealed {
		return errors.New("spool is sealed for reading")
	}
	record, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("encode spooled row %q: %w", row.ID, err)
	}
	record = append(record, '\n')
	estimatedBytes := int64(len(record) + 24)

	if spool.file == nil && spool.memoryUsed+estimatedBytes <= spool.memoryLimit {
		spool.records = append(spool.records, record)
		spool.memoryUsed += estimatedBytes
		return nil
	}
	if spool.file == nil {
		if err := spool.startDiskSpool(); err != nil {
			return err
		}
	}
	if _, err := spool.writer.Write(record); err != nil {
		return fmt.Errorf("write spooled row %q: %w", row.ID, err)
	}
	return nil
}

func (spool *BoundedSpool) Stream() (RowStream, error) {
	if spool.closed {
		return nil, errors.New("spool is closed")
	}
	if spool.sealed {
		return nil, errors.New("spool stream was already opened")
	}
	spool.sealed = true

	if spool.file == nil {
		stream := &encodedRowStream{records: spool.records}
		spool.active = stream
		return stream, nil
	}
	if err := spool.writer.Flush(); err != nil {
		return nil, fmt.Errorf("flush disk spool: %w", err)
	}
	if err := spool.file.Sync(); err != nil {
		return nil, fmt.Errorf("sync disk spool: %w", err)
	}
	if err := spool.file.Close(); err != nil {
		return nil, fmt.Errorf("close disk spool writer: %w", err)
	}
	spool.file = nil
	spool.writer = nil

	reader, err := os.Open(spool.path)
	if err != nil {
		return nil, fmt.Errorf("open disk spool: %w", err)
	}
	stream := &fileRowStream{file: reader, decoder: json.NewDecoder(reader)}
	spool.active = stream
	return stream, nil
}

func (spool *BoundedSpool) Spilled() bool {
	return spool.path != ""
}

func (spool *BoundedSpool) Close() error {
	if spool.closed {
		return nil
	}
	spool.closed = true
	var closeError error
	if spool.active != nil {
		closeError = spool.active.Close()
	}
	if spool.writer != nil {
		if err := spool.writer.Flush(); err != nil && closeError == nil {
			closeError = err
		}
	}
	if spool.file != nil {
		if err := spool.file.Close(); err != nil && closeError == nil {
			closeError = err
		}
	}
	if spool.path != "" {
		if err := os.Remove(spool.path); err != nil && !errors.Is(err, os.ErrNotExist) && closeError == nil {
			closeError = err
		}
	}
	spool.records = nil
	return closeError
}

func (spool *BoundedSpool) startDiskSpool() error {
	if err := os.MkdirAll(spool.directory, 0o700); err != nil {
		return fmt.Errorf("create spool directory: %w", err)
	}
	file, err := os.CreateTemp(spool.directory, "rows-*.jsons")
	if err != nil {
		return fmt.Errorf("create disk spool: %w", err)
	}
	spool.file = file
	spool.path, err = filepath.Abs(file.Name())
	if err != nil {
		file.Close()
		return fmt.Errorf("resolve disk spool path: %w", err)
	}
	spool.writer = bufio.NewWriter(file)
	for _, record := range spool.records {
		if _, err := spool.writer.Write(record); err != nil {
			return fmt.Errorf("flush memory rows to disk spool: %w", err)
		}
	}
	spool.records = nil
	spool.memoryUsed = 0
	return nil
}

type encodedRowStream struct {
	records [][]byte
	index   int
	closed  bool
}

func (stream *encodedRowStream) Next(ctx context.Context) (Row, error) {
	if stream.closed {
		return Row{}, errors.New("row stream is closed")
	}
	if err := ctx.Err(); err != nil {
		return Row{}, err
	}
	if stream.index >= len(stream.records) {
		return Row{}, io.EOF
	}
	var row Row
	if err := json.NewDecoder(bytes.NewReader(stream.records[stream.index])).Decode(&row); err != nil {
		return Row{}, fmt.Errorf("decode memory-spooled row: %w", err)
	}
	stream.index++
	return row, nil
}

func (stream *encodedRowStream) Close() error {
	stream.closed = true
	return nil
}

type fileRowStream struct {
	file    *os.File
	decoder *json.Decoder
	closed  bool
}

func (stream *fileRowStream) Next(ctx context.Context) (Row, error) {
	if stream.closed {
		return Row{}, errors.New("row stream is closed")
	}
	if err := ctx.Err(); err != nil {
		return Row{}, err
	}
	var row Row
	if err := stream.decoder.Decode(&row); err != nil {
		return Row{}, err
	}
	return row, nil
}

func (stream *fileRowStream) Close() error {
	if stream.closed {
		return nil
	}
	stream.closed = true
	return stream.file.Close()
}
