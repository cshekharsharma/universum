package compression

import (
	"testing"
	"universum/config"
)

func setupCompressionTests() {
	config.Store = config.GetSkeleton()
}

func TestGetCompressorWithOptions(t *testing.T) {
	setupCompressionTests()
	opts := &Options{
		CompressionAlgo: CompressionAlgoLZ4,
	}

	compressor := GetCompressor(opts)
	if compressor == nil {
		t.Fatalf("Expected compressor, got nil")
	}

	_, isLZ4 := compressor.(*LZ4Compressor)
	if !isLZ4 {
		t.Fatalf("Expected LZ4Compressor, got %T", compressor)
	}
}

func TestGetCompressorWithConfig(t *testing.T) {
	setupCompressionTests()
	config.Store.Storage.Memory.SnapshotCompressionAlgo = config.CompressionAlgoLZ4

	opts := &Options{
		CompressionAlgo: "",
	}

	compressor := GetCompressor(opts)
	if compressor == nil {
		t.Fatalf("Expected compressor, got nil")
	}

	_, isNoCompression := compressor.(*NoCompressor)
	if !isNoCompression {
		t.Fatalf("Expected NoCompressor from config, got %T", compressor)
	}
}

func TestGetCompressorDefaultCase(t *testing.T) {
	setupCompressionTests()

	opts := &Options{
		CompressionAlgo: "UnknownAlgorithm",
	}

	compressor := GetCompressor(opts)
	if compressor == nil {
		t.Fatalf("Expected compressor, got nil")
	}

	_, isLZ4 := compressor.(*LZ4Compressor)
	if !isLZ4 {
		t.Fatalf("Expected default LZ4Compressor, got %T", compressor)
	}
}
