package testsupport

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const scaleJSONPaddingBytes = 16 * 1024

// ScaleFixture describes a deterministic generated input. Scale fixtures are
// created only in disposable test directories; the large files are never
// committed to the repository.
type ScaleFixture struct {
	Path string
	Rows int
	Size int64
}

// GenerateJSONScaleFixture writes a top-level Data array until the completed
// JSON document is at least targetBytes. The overshoot is bounded by one row.
func GenerateJSONScaleFixture(path string, targetBytes int64) (ScaleFixture, error) {
	const prefix = `{"Data":[`
	const suffix = `]}`
	if targetBytes <= int64(len(prefix)+len(suffix)) {
		return ScaleFixture{}, errors.New("JSON scale target is too small")
	}

	file, absolutePath, err := createScaleFixture(path)
	if err != nil {
		return ScaleFixture{}, err
	}
	completed := false
	defer func() {
		_ = file.Close()
		if !completed {
			_ = os.Remove(absolutePath)
		}
	}()

	writer := bufio.NewWriterSize(file, 256*1024)
	if _, err := writer.WriteString(prefix); err != nil {
		return ScaleFixture{}, fmt.Errorf("write JSON scale prefix: %w", err)
	}
	written := int64(len(prefix))
	padding := strings.Repeat("x", scaleJSONPaddingBytes)
	rows := 0
	for written+int64(len(suffix)) < targetBytes {
		ordinal := rows + 1
		record, err := json.Marshal(scaleJSONRecord{
			ID: ordinal, Name: scaleRecordName(ordinal), City: scaleRecordCity(ordinal),
			Marker: "עברית, English 🙂", Padding: padding,
		})
		if err != nil {
			return ScaleFixture{}, fmt.Errorf("encode JSON scale row %d: %w", ordinal, err)
		}
		if rows > 0 {
			if err := writer.WriteByte(','); err != nil {
				return ScaleFixture{}, fmt.Errorf("write JSON scale separator: %w", err)
			}
			written++
		}
		if _, err := writer.Write(record); err != nil {
			return ScaleFixture{}, fmt.Errorf("write JSON scale row %d: %w", ordinal, err)
		}
		written += int64(len(record))
		rows++
	}
	if _, err := writer.WriteString(suffix); err != nil {
		return ScaleFixture{}, fmt.Errorf("write JSON scale suffix: %w", err)
	}
	if err := finishScaleFixture(file, writer); err != nil {
		return ScaleFixture{}, err
	}
	completed = true
	return scaleFixtureInfo(absolutePath, rows)
}

// GenerateCSVScaleFixture writes exactly rows deterministic UTF-8 records.
func GenerateCSVScaleFixture(path string, rows int) (ScaleFixture, error) {
	if rows <= 0 {
		return ScaleFixture{}, errors.New("CSV scale row count must be greater than zero")
	}

	file, absolutePath, err := createScaleFixture(path)
	if err != nil {
		return ScaleFixture{}, err
	}
	completed := false
	defer func() {
		_ = file.Close()
		if !completed {
			_ = os.Remove(absolutePath)
		}
	}()

	buffer := bufio.NewWriterSize(file, 256*1024)
	writer := csv.NewWriter(buffer)
	if err := writer.Write([]string{"id", "name", "city", "marker"}); err != nil {
		return ScaleFixture{}, fmt.Errorf("write CSV scale header: %w", err)
	}
	for ordinal := 1; ordinal <= rows; ordinal++ {
		if err := writer.Write([]string{
			strconv.Itoa(ordinal), scaleRecordName(ordinal), scaleRecordCity(ordinal), "עברית, English 🙂",
		}); err != nil {
			return ScaleFixture{}, fmt.Errorf("write CSV scale row %d: %w", ordinal, err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return ScaleFixture{}, fmt.Errorf("flush CSV scale rows: %w", err)
	}
	if err := finishScaleFixture(file, buffer); err != nil {
		return ScaleFixture{}, err
	}
	completed = true
	return scaleFixtureInfo(absolutePath, rows)
}

type scaleJSONRecord struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	City    string `json:"city"`
	Marker  string `json:"marker"`
	Padding string `json:"padding"`
}

func scaleRecordName(ordinal int) string {
	return fmt.Sprintf("רשומה-%06d 🧭", ordinal)
}

func scaleRecordCity(ordinal int) string {
	if ordinal%2 == 1 {
		return "בת ים"
	}
	return "תל אביב"
}

func createScaleFixture(path string) (*os.File, string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, "", fmt.Errorf("resolve scale fixture path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return nil, "", fmt.Errorf("create scale fixture directory: %w", err)
	}
	file, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, "", fmt.Errorf("create scale fixture %q: %w", absolutePath, err)
	}
	return file, absolutePath, nil
}

func finishScaleFixture(file *os.File, writer *bufio.Writer) error {
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush scale fixture: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync scale fixture: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close scale fixture: %w", err)
	}
	return nil
}

func scaleFixtureInfo(path string, rows int) (ScaleFixture, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ScaleFixture{}, fmt.Errorf("stat scale fixture: %w", err)
	}
	return ScaleFixture{Path: path, Rows: rows, Size: info.Size()}, nil
}
