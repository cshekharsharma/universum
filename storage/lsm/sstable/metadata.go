package sstable

type Metadata struct {
	Version           int    // Version of the SSTable format
	NumRecords        int    // Number of key-value pairs in the SSTable
	IndexOffset       int64  // Offset in the file where the index block starts
	DataSize          int64  // Total size of the data block
	IndexSize         int64  // Size of the index block
	BloomFilterOffset int64  // Offset for the bloom filter block (if used)
	BloomFilterSize   int64  // Size of the bloom filter block
	Timestamp         int64  // Timestamp when this SSTable was created
	Compression       string // Compression algorithm used (if any)
	Checksum          uint32 // Optional checksum for data validation
}
