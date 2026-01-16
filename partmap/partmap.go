package partmap

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/twmb/murmur3"
)

type (
	PartitionRangeStart uint32
	ServerIndex         uint16
)

// PartitionMapping is map of partition number(the beginning of the partition range in the 32-bit space) to server index
type PartitionMapping map[PartitionRangeStart]ServerIndex

// MarshalToRawBytes marshalls PartitionMapping to raw bytes
func MarshalToRawBytes(pm PartitionMapping) ([]byte, error) {
	var compressedData bytes.Buffer

	listLength := uint32(len(pm))
	err := binary.Write(&compressedData, binary.BigEndian, listLength)
	if err != nil {
		return nil, fmt.Errorf("failed to write list length: %w", err)
	}

	// Write each pair of integers to the compressed data buffer
	for partition, serverIdx := range pm {
		err := binary.Write(&compressedData, binary.BigEndian, partition)
		if err != nil {
			return nil, fmt.Errorf("failed to write partition range start: %w", err)
		}
		err = binary.Write(&compressedData, binary.BigEndian, serverIdx)
		if err != nil {
			return nil, fmt.Errorf("failed to write server index: %w", err)
		}
	}

	// Compress the data using gzip
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)
	_, err = gzipWriter.Write(compressedData.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return compressedBuffer.Bytes(), nil
}

// UnmarshalToPartitionMapping unmarshalls raw bytes to PartitionMapping
func UnmarshalToPartitionMapping(data []byte) (PartitionMapping, error) {
	decompressed, err := decompressGzip(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	mapping, err := parseBinaryFormat(decompressed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse binary format: %w", err)
	}

	return mapping, nil
}

// Clone creates a copy of the PartitionMapping
func (pm PartitionMapping) Clone() PartitionMapping {
	if pm == nil {
		return nil
	}

	clone := make(PartitionMapping, len(pm))
	for k, v := range pm {
		clone[k] = v
	}
	return clone
}

// Murmur3Partition32 computes the partition index and beginning of the partition range in a 32-bit space for a given key and number of partitions.
//
// Please note that numPartitions must be a power of 2 (e.g. 1, 2, 4, 8, 16, ...), otherwise the distribution may not be uniform and could lead to unexpected results, including panics
//
// Implementation details:
//
//   - It uses the Murmur3 hashing algorithm to generate a 32-bit hash of the key, then calculates the partition index by dividing the hash value by the size of each partition window.
//   - It returns the partition index and the corresponding partition value (the start of the partition range).
func Murmur3Partition32(key string, numPartitions uint32) (uint32, uint32) {
	if numPartitions == 1 {
		return 0, 0 // Return zero values if numPartitions is 1 (as there's only one partition)
	}
	window := uint32((1 << 32) / int(numPartitions))    // Size of each partition window (dividing the full 32-bit range by number of partitions)
	partitionIdx := murmur3.Sum32([]byte(key)) / window // Determine partition index by dividing hash by window size
	partition := partitionIdx * window                  // Calculate the start of the partition range
	return partitionIdx, partition
}

// decompressGzip decompresses gzipped data
func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating gzip reader: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading decompressed data: %w", err)
	}

	return decompressed, nil
}

// parseBinaryFormat parses the binary format
func parseBinaryFormat(data []byte) (PartitionMapping, error) {
	var listLength uint32
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.BigEndian, &listLength)
	if err != nil {
		return nil, fmt.Errorf("reading list length: %w", err)
	}

	result := make(PartitionMapping)
	for i := uint32(0); i < listLength; i++ {
		var (
			partition PartitionRangeStart
			serverIdx ServerIndex
		)
		if err := binary.Read(reader, binary.BigEndian, &partition); err != nil {
			return nil, fmt.Errorf("reading int value: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &serverIdx); err != nil {
			return nil, fmt.Errorf("reading short value: %w", err)
		}

		result[partition] = serverIdx
	}

	return result, nil
}
