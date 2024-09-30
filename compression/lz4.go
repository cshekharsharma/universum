package compression

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/pierrec/lz4"
)

type LZ4Compressor struct {
	reader   io.Reader
	writer   io.Writer
	options  *Options
	lzReader *lz4.Reader
	lzWriter *lz4.Writer
}

// SetOptions sets the options for the compressor
func (c *LZ4Compressor) Init(opts *Options) {
	if opts == nil {
		return
	}

	c.options = opts

	if c.options.Reader != nil {
		c.reader = c.options.Reader
		c.lzReader = lz4.NewReader(c.reader)
	}

	if c.options.Writer != nil {
		c.writer = c.options.Writer
		c.lzWriter = lz4.NewWriter(c.writer)
	}
}

// CompressAndWrite compresses the data and writes it to the io.Writer
// if no io.Writer is set, it will return an error
// it will optionally close the internal writer if AutoCloseWriter is set to true
// if AutoCloseWriter is set to false, the caller is responsible for closing the writer
// using compressor.Close() function
func (c *LZ4Compressor) CompressAndWrite(data []byte) error {
	if c.lzWriter == nil {
		return errors.New("no destination io.writer provided for compression")
	}

	_, err := c.lzWriter.Write(data)
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

// DecompressAndRead reads the compressed data from the io.Reader and decompresses it
// if no io.Reader is set, it will return an error
// it will return the decompressed data as byte array
func (c *LZ4Compressor) DecompressAndRead(chunkSize int64) ([]byte, error) {
	if c.lzReader == nil {
		return nil, errors.New("no source io.reader provided for decompression")
	}

	buf := make([]byte, chunkSize)

	n, err := c.lzReader.Read(buf)
	if err != nil {
		if err == io.EOF {
			if n > 0 {
				return buf[:n], nil
			}
			return nil, io.EOF // End of file reached
		}
		return nil, fmt.Errorf("error reading compressed data: %v", err)
	}

	return buf[:n], nil
}

// Compress is a stateless compression function that compresses data using LZ4
// it takes input as byte array and returns the compressed byte array
// if there is any preset io.Writer set to this compressor, it will be ignored.
func (c *LZ4Compressor) Compress(data []byte) ([]byte, error) {
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

// Decompress is a stateless decompression function that decompresses data using LZ4
// it takes input as byte array and returns the decompressed byte array
// if there is any preset io.Reader set to this compressor, it will be ignored.
func (c *LZ4Compressor) Decompress(data []byte) ([]byte, error) {
	var decompressed bytes.Buffer
	lzReader := lz4.NewReader(bytes.NewReader(data))

	if _, err := io.Copy(&decompressed, lzReader); err != nil {
		return nil, err
	}

	return decompressed.Bytes(), nil
}

// Close closes the writer and other compressor resources
func (c *LZ4Compressor) Close() error {
	if c.lzWriter != nil {
		err := c.lzWriter.Close()
		if err != nil {
			return err
		}
		c.lzWriter = nil
	}

	return nil
}
