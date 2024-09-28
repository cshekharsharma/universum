package compression

import "io"

type Options struct {
	Reader          io.Reader
	Writer          io.Writer
	CompressionAlgo CompressionAlgo
	AutoCloseWriter bool
}
