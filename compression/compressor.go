package compression

import (
	"universum/config"
)

type CompressionAlgo string

const (
	CompressionAlgoNone CompressionAlgo = "NONE"
	CompressionAlgoLZ4  CompressionAlgo = "LZ4"
)

type Compressor interface {
	Init(opts *Options)
	Compress(data []byte) ([]byte, error)
	CompressAndWrite(data []byte) error
	Decompress(data []byte) ([]byte, error)
	DecompressAndRead(chunkSize int64) ([]byte, error)
	Close() error
}

func GetCompressor(opts *Options) Compressor {
	algo := config.Store.Storage.Memory.SnapshotCompressionAlgo
	if opts.CompressionAlgo != "" {
		algo = string(opts.CompressionAlgo)
	}

	switch algo {
	case config.CompressionAlgoLZ4:
		c := &LZ4Compressor{}
		c.Init(opts)
		return c

	case config.CompressionAlgoNone:
		c := &NoCompressor{}
		c.Init(opts)
		return c

	default:
		c := &LZ4Compressor{}
		c.Init(opts)
		return c
	}
}

func IsCompressionEnabled() bool {
	return config.Store.Storage.Memory.SnapshotCompressionAlgo != config.CompressionAlgoNone
}
