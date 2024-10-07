package compression

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNoCompressorInit(t *testing.T) {
	compressor := &NoCompressor{}
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
}

func TestNoCompressorCompressAndWrite(t *testing.T) {
	data := []byte("test data")
	buffer := &bytes.Buffer{}
	opts := &Options{
		Writer:          buffer,
		AutoCloseWriter: false,
	}

	compressor := &NoCompressor{}
	compressor.Init(opts)

	err := compressor.CompressAndWrite(data)
	if err != nil {
		t.Fatalf("CompressAndWrite failed: %v", err)
	}

	if buffer.String() != string(data) {
		t.Errorf("Expected buffer to contain '%s', got '%s'", data, buffer.String())
	}
}

func TestNoCompressorCompressAndWriteNoWriter(t *testing.T) {
	data := []byte("test data")
	compressor := &NoCompressor{}
	compressor.Init(&Options{})

	err := compressor.CompressAndWrite(data)
	if err == nil {
		t.Fatalf("Expected error when writer is nil")
	}

	expectedErr := "no destination io.Writer provided"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestNoCompressorCompressAndWriteAutoClose(t *testing.T) {
	data := []byte("test data")
	buffer := &bytes.Buffer{}
	opts := &Options{
		Writer:          buffer,
		AutoCloseWriter: true,
	}

	compressor := &NoCompressor{}
	compressor.Init(opts)

	err := compressor.CompressAndWrite(data)
	if err != nil {
		t.Fatalf("CompressAndWrite failed: %v", err)
	}

	if buffer.String() != string(data) {
		t.Errorf("Expected buffer to contain '%s', got '%s'", data, buffer.String())
	}
}

func TestNoCompressorDecompressAndRead(t *testing.T) {
	data := []byte("test data")
	reader := strings.NewReader(string(data))
	opts := &Options{
		Reader: reader,
	}

	compressor := &NoCompressor{}
	compressor.Init(opts)

	chunkSize := int64(len(data))
	readData, err := compressor.DecompressAndRead(chunkSize)
	if err != nil && err != io.EOF {
		t.Fatalf("DecompressAndRead failed: %v", err)
	}

	if !bytes.Equal(readData, data) {
		t.Errorf("Expected read data to be '%s', got '%s'", data, readData)
	}
}

func TestNoCompressorDecompressAndReadNoReader(t *testing.T) {
	compressor := &NoCompressor{}
	compressor.Init(&Options{})

	_, err := compressor.DecompressAndRead(1024)
	if err == nil {
		t.Fatalf("Expected error when reader is nil")
	}

	expectedErr := "no source io.Reader provided"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestNoCompressorCompress(t *testing.T) {
	data := []byte("test data")
	compressor := &NoCompressor{}

	compressedData, err := compressor.Compress(data)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if !bytes.Equal(compressedData, data) {
		t.Errorf("Expected compressed data to be '%s', got '%s'", data, compressedData)
	}
}

func TestNoCompressorDecompress(t *testing.T) {
	data := []byte("test data")
	compressor := &NoCompressor{}

	decompressedData, err := compressor.Decompress(data)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(decompressedData, data) {
		t.Errorf("Expected decompressed data to be '%s', got '%s'", data, decompressedData)
	}
}

func TestNoCompressorClose(t *testing.T) {
	compressor := &NoCompressor{}
	err := compressor.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestNoCompressorDecompressAndReadEOF(t *testing.T) {
	data := []byte("test data")
	reader := strings.NewReader(string(data))
	opts := &Options{
		Reader: reader,
	}

	compressor := &NoCompressor{}
	compressor.Init(opts)

	chunkSize := int64(len(data) + 10) // Larger than data length
	readData, err := compressor.DecompressAndRead(chunkSize)
	if err != nil && err != io.EOF {
		t.Fatalf("DecompressAndRead failed: %v", err)
	}

	if !bytes.Equal(readData, data) {
		t.Errorf("Expected read data to be '%s', got '%s'", data, readData)
	}
}

func TestNoCompressorCompressAndWriteError(t *testing.T) {
	data := []byte("test data")
	writer := &ErrorWriter{}
	opts := &Options{
		Writer:          writer,
		AutoCloseWriter: false,
	}

	compressor := &NoCompressor{}
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

type ErrorWriter struct{}

func (w *ErrorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("simulated write error")
}
