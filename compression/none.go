package compression

import "errors"

type NilCompressor struct {
	options *Options
}

func (c *NilCompressor) Init(opts *Options) {
	c.options = opts
}

func (c *NilCompressor) Compress(data []byte) ([]byte, error) {
	return nil, errors.New("no compressor selected")
}

func (c *NilCompressor) CompressAndWrite(data []byte) error {
	return errors.New("no compressor selected")
}

func (c *NilCompressor) Decompress(data []byte) ([]byte, error) {
	return nil, errors.New("no compressor selected")
}

func (c *NilCompressor) Close() error {
	return errors.New("no compressor selected")
}

func (c *NilCompressor) DecompressAndRead(chunkSize int64) ([]byte, error) {
	return nil, errors.New("no compressor selected")
}
