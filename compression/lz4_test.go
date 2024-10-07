package compression

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/pierrec/lz4"
)

func TestLZ4CompressorInit(t *testing.T) {
	compressor := &LZ4Compressor{}
	reader := strings.NewReader("test data")
	writer := &bytes.Buffer{}
	opts := &Options{
		Reader: reader,
		Writer: writer,
	}

	compressor.Init(opts)

	if compressor.reader != reader {
		t.Errorf("Expected reader to be set")
	}

	if compressor.writer != writer {
		t.Errorf("Expected writer to be set")
	}

	if compressor.options != opts {
		t.Errorf("Expected options to be set")
	}

	if compressor.lzReader == nil {
		t.Errorf("Expected lzReader to be initialized")
	}

	if compressor.lzWriter == nil {
		t.Errorf("Expected lzWriter to be initialized")
	}
}

func TestLZ4CompressorCompressAndWrite(t *testing.T) {
	data := []byte("test data for compression")
	buffer := &bytes.Buffer{}
	opts := &Options{
		Writer:          buffer,
		AutoCloseWriter: false,
	}

	compressor := &LZ4Compressor{}
	compressor.Init(opts)

	err := compressor.CompressAndWrite(data)
	if err != nil {
		t.Fatalf("CompressAndWrite failed: %v", err)
	}

	if buffer.Len() == 0 {
		t.Errorf("Expected compressed data to be written to buffer")
	}
}

func TestLZ4CompressorCompressAndWriteNoWriter(t *testing.T) {
	data := []byte("test data")
	compressor := &LZ4Compressor{}
	compressor.Init(&Options{})

	err := compressor.CompressAndWrite(data)
	if err == nil {
		t.Fatalf("Expected error when writer is nil")
	}

	expectedErr := "no destination io.writer provided for compression"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestLZ4CompressorCompressAndWriteAutoClose(t *testing.T) {
	data := []byte("test data for compression")
	buffer := &bytes.Buffer{}
	opts := &Options{
		Writer:          buffer,
		AutoCloseWriter: true,
	}

	compressor := &LZ4Compressor{}
	compressor.Init(opts)

	err := compressor.CompressAndWrite(data)
	if err != nil {
		t.Fatalf("CompressAndWrite failed: %v", err)
	}

	if compressor.lzWriter != nil {
		t.Errorf("Expected lzWriter to be nil after Close")
	}

	if buffer.Len() == 0 {
		t.Errorf("Expected compressed data to be written to buffer")
	}
}

func TestLZ4CompressorDecompressAndRead(t *testing.T) {
	data := []byte("test data for decompression")
	compressedData, err := compressLZ4Data(data)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	reader := bytes.NewReader(compressedData)
	opts := &Options{
		Reader: reader,
	}

	compressor := &LZ4Compressor{}
	compressor.Init(opts)

	chunkSize := int64(len(data) + 100) // Larger to read all data
	decompressedData, err := compressor.DecompressAndRead(chunkSize)
	if err != nil && err != io.EOF {
		t.Fatalf("DecompressAndRead failed: %v", err)
	}

	if !bytes.Equal(decompressedData, data) {
		t.Errorf("Expected decompressed data to be '%s', got '%s'", data, decompressedData)
	}
}

func compressLZ4Data(data []byte) ([]byte, error) {
	var compressed bytes.Buffer
	lzWriter := lz4.NewWriter(&compressed)
	if _, err := lzWriter.Write(data); err != nil {
		return nil, err
	}
	if err := lzWriter.Close(); err != nil {
		return nil, err
	}
	return compressed.Bytes(), nil
}

func TestLZ4CompressorDecompressAndReadNoReader(t *testing.T) {
	compressor := &LZ4Compressor{}
	compressor.Init(&Options{})

	_, err := compressor.DecompressAndRead(1024)
	if err == nil {
		t.Fatalf("Expected error when reader is nil")
	}

	expectedErr := "no source io.reader provided for decompression"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestLZ4CompressorCompress(t *testing.T) {
	data := []byte("test data for stateless compression")
	compressor := &LZ4Compressor{}

	_, err := compressor.Compress(data)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
}

func TestLZ4CompressorDecompress(t *testing.T) {
	data := []byte("test data for stateless decompression")
	compressedData, err := compressLZ4Data(data)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	compressor := &LZ4Compressor{}
	decompressedData, err := compressor.Decompress(compressedData)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(decompressedData, data) {
		t.Errorf("Expected decompressed data to be '%s', got '%s'", data, decompressedData)
	}
}

func TestLZ4CompressorClose(t *testing.T) {
	buffer := &bytes.Buffer{}
	opts := &Options{
		Writer: buffer,
	}

	compressor := &LZ4Compressor{}
	compressor.Init(opts)

	err := compressor.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if compressor.lzWriter != nil {
		t.Errorf("Expected lzWriter to be nil after Close")
	}
}

func TestLZ4CompressorCompressAndWriteError(t *testing.T) {
	data := []byte("test data")
	writer := &ErrorWriter{}
	opts := &Options{
		Writer:          writer,
		AutoCloseWriter: false,
	}

	compressor := &LZ4Compressor{}
	compressor.Init(opts)

	err := compressor.CompressAndWrite(data)
	if err == nil {
		t.Fatalf("Expected error when writer fails")
	}

	expectedErr := "simulated write error"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestLZ4CompressorDecompressAndReadError(t *testing.T) {
	reader := &ErrorReader{}
	opts := &Options{
		Reader: reader,
	}

	compressor := &LZ4Compressor{}
	compressor.Init(opts)

	_, err := compressor.DecompressAndRead(1024)
	if err == nil {
		t.Fatalf("Expected error when reader fails")
	}

	expectedErr := "simulated read error"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedErr, err.Error())
	}
}

type ErrorReader struct{}

func (r *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}
