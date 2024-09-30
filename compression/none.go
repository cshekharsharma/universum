package compression

import (
	"errors"
	"io"
)

type NoCompressor struct {
	reader  io.Reader
	writer  io.Writer
	options *Options
}

// SetOptions sets the options for the NoCompressor
func (c *NoCompressor) Init(opts *Options) {
	if opts == nil {
		return
	}

	c.options = opts

	if c.options.Reader != nil {
		c.reader = c.options.Reader
	}

	if c.options.Writer != nil {
		c.writer = c.options.Writer
	}
}

// CompressAndWrite writes the data to the io.Writer without compression
// if no io.Writer is set, it will return an error
// it will optionally close the internal writer if AutoCloseWriter is set to true
// if AutoCloseWriter is set to false, the caller is responsible for closing the writer
func (c *NoCompressor) CompressAndWrite(data []byte) error {
	if c.writer == nil {
		return errors.New("no destination io.Writer provided")
	}

	_, err := c.writer.Write(data)
	if err != nil {
		return err
	}

	if c.options.AutoCloseWriter {
		if err := c.Close(); err != nil {
			return err
		}
	}

	return nil
}

// DecompressAndRead reads the uncompressed data from the io.Reader
// if no io.Reader is set, it will return an error
// it will return the uncompressed data as a byte array
func (c *NoCompressor) DecompressAndRead(chunkSize int64) ([]byte, error) {
	if c.reader == nil {
		return nil, errors.New("no source io.Reader provided")
	}

	buf := make([]byte, chunkSize)
	n, err := c.reader.Read(buf)
	if err != nil {
		if err == io.EOF {
			if n > 0 {
				return buf[:n], nil
			}
			return nil, io.EOF
		}
		return nil, err
	}

	return buf[:n], nil
}

// Compress returns the data as-is without any compression
func (c *NoCompressor) Compress(data []byte) ([]byte, error) {
	// No compression, just return the original data
	return data, nil
}

// Decompress returns the data as-is without any decompression
func (c *NoCompressor) Decompress(data []byte) ([]byte, error) {
	// No decompression, just return the original data
	return data, nil
}

// Close is a no-op for the NoCompressor but included to match the interface
func (c *NoCompressor) Close() error {
	return nil
}
